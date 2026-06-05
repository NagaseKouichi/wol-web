package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "server/migrations"

	wol "github.com/HuakunShen/wol/wol-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), true))
		se.Router.GET("/api/hosts", func(e *core.RequestEvent) error {
			return e.JSON(200, map[string]string{"message": "Host waked"})
		})
		se.Router.POST("/api/host-statuses", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return e.JSON(401, map[string]string{"message": "Unauthorized"})
			}

			info, err := e.RequestInfo()
			if err != nil {
				return e.BadRequestError("Failed to read request info", err)
			}

			rawIds, ok := info.Body["ids"].([]any)
			if !ok {
				return e.BadRequestError("Failed to read host ids from body", nil)
			}

			statuses := map[string]hostStatus{}
			for _, rawId := range rawIds {
				id, ok := rawId.(string)
				if !ok {
					continue
				}

				requestedHost, err := app.FindRecordById("hosts", id)
				if err != nil || requestedHost.GetString("user") != e.Auth.Id {
					continue
				}

				status := resolveHostStatus(
					requestedHost.GetString("mac"),
					requestedHost.GetString("ip"),
					requestedHost.GetString("hostIp"),
					requestedHost.GetInt("port"),
				)
				status.PowerAvailable = strings.TrimSpace(requestedHost.GetString("agentUrl")) != "" &&
					strings.TrimSpace(requestedHost.GetString("agentToken")) != ""
				statuses[id] = status
			}

			return e.JSON(200, statuses)
		})
		se.Router.POST("/api/host-config", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return e.JSON(401, map[string]string{"message": "Unauthorized"})
			}

			var data hostConfigRequest
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}

			record, err := upsertHostConfig(app, e.Auth.Id, data)
			if err != nil {
				return e.BadRequestError("Failed to save host config", err)
			}

			return e.JSON(200, map[string]string{
				"message": "Host config saved",
				"id":      record.Id,
			})
		})
		se.Router.POST("/api/host-power", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return e.JSON(401, map[string]string{"message": "Unauthorized"})
			}

			info, err := e.RequestInfo()
			if err != nil {
				return e.BadRequestError("Failed to read request info", err)
			}

			id, ok := info.Body["id"].(string)
			if !ok {
				return e.BadRequestError("Failed to read id from body", nil)
			}

			action, ok := info.Body["action"].(string)
			if !ok || (action != "shutdown" && action != "sleep") {
				return e.BadRequestError("Invalid power action", nil)
			}

			requestedHost, err := app.FindRecordById("hosts", id)
			if err != nil {
				return e.BadRequestError("Failed to find host", err)
			}
			if requestedHost.GetString("user") != e.Auth.Id {
				return e.JSON(401, map[string]string{"message": "Unauthorized"})
			}

			agentURL := requestedHost.GetString("agentUrl")
			agentToken := requestedHost.GetString("agentToken")
			if strings.TrimSpace(agentURL) == "" || strings.TrimSpace(agentToken) == "" {
				return e.BadRequestError("Host agent is not configured", nil)
			}

			if err := callHostAgent(agentURL, agentToken, action); err != nil {
				return e.JSON(502, map[string]string{
					"message": "Failed to call host agent",
					"error":   err.Error(),
				})
			}

			return e.JSON(200, map[string]string{"message": "Power action requested"})
		})
		se.Router.POST("/api/wake", func(e *core.RequestEvent) error {
			data := struct {
				// unexported to prevent binding
				Id string `json:"id" form:"id"`
			}{}
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}
			info, err := e.RequestInfo()
			if err != nil {
				return e.BadRequestError("Failed to read request info", err)
			}
			id, ok := info.Body["id"].(string)
			if !ok {
				return e.BadRequestError("Failed to read id from body", err)
			}

			isAuthenticated := e.Auth != nil
			if !isAuthenticated {
				return e.JSON(401, map[string]string{"message": "Unauthorized"})
			}
			userId := e.Auth.Id
			requestedHost, err := app.FindRecordById("hosts", id)
			if err != nil {
				return e.BadRequestError("Failed to find host", err)
			}
			// Get the user field - it will return an interface{} that you can type assert
			recordOwnerId := requestedHost.GetString("user")
			if recordOwnerId != userId {
				return e.JSON(401, map[string]string{"message": "Unauthorized"})
			}
			mac := requestedHost.GetString("mac")
			port := requestedHost.GetInt("port")
			targetIp := requestedHost.GetString("ip")
			err = wol.WakeOnLan(mac, targetIp, strconv.Itoa(port))
			if err != nil {
				return e.JSON(400, map[string]string{"message": "Failed to wake host", "error": err.Error()})
			}
			return e.JSON(200, map[string]string{"message": "WakeOnLan Magic Packet Sent"})
		})
		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

type hostStatus struct {
	Online         bool   `json:"online"`
	IP             string `json:"ip,omitempty"`
	PowerAvailable bool   `json:"powerAvailable"`
}

type hostConfigRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Mac        string `json:"mac"`
	IP         string `json:"ip"`
	HostIP     string `json:"hostIp"`
	Port       int    `json:"port"`
	AgentURL   string `json:"agentUrl"`
	AgentToken string `json:"agentToken"`
}

func upsertHostConfig(app *pocketbase.PocketBase, userID string, data hostConfigRequest) (*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("hosts")
	if err != nil {
		return nil, err
	}

	var record *core.Record
	if strings.TrimSpace(data.ID) == "" {
		record = core.NewRecord(collection)
		record.Set("user", userID)
	} else {
		record, err = app.FindRecordById("hosts", data.ID)
		if err != nil {
			return nil, err
		}
		if record.GetString("user") != userID {
			return nil, errors.New("unauthorized")
		}
	}

	record.Set("name", strings.TrimSpace(data.Name))
	record.Set("mac", strings.TrimSpace(data.Mac))
	record.Set("ip", strings.TrimSpace(data.IP))
	record.Set("hostIp", strings.TrimSpace(data.HostIP))
	record.Set("port", data.Port)
	record.Set("agentUrl", strings.TrimSpace(data.AgentURL))
	if strings.TrimSpace(data.AgentToken) != "" {
		record.Set("agentToken", strings.TrimSpace(data.AgentToken))
	}

	if err := app.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

func resolveHostStatus(macAddress string, broadcastAddress string, hostIP string, port int) hostStatus {
	hostIP = strings.TrimSpace(hostIP)
	if hostIP != "" {
		return hostStatus{
			Online: pingHost(hostIP),
			IP:     hostIP,
		}
	}

	normalizedMAC := normalizeMAC(macAddress)
	if normalizedMAC == "" {
		return hostStatus{Online: false}
	}

	if ip, ok := findARPIPByMAC(normalizedMAC); ok {
		return hostStatus{Online: true, IP: ip}
	}

	probeSubnet(broadcastAddress, port)
	time.Sleep(300 * time.Millisecond)

	if ip, ok := findARPIPByMAC(normalizedMAC); ok {
		return hostStatus{Online: true, IP: ip}
	}

	return hostStatus{Online: false}
}

func pingHost(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", parsedIP.String())
	return command.Run() == nil
}

func callHostAgent(agentURL string, token string, action string) error {
	agentURL = strings.TrimRight(strings.TrimSpace(agentURL), "/")
	token = strings.TrimSpace(token)

	body, err := json.Marshal(map[string]string{"action": action})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, agentURL+"/api/power", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	return nil
}

func normalizeMAC(value string) string {
	mac, err := net.ParseMAC(value)
	if err != nil {
		return ""
	}
	return strings.ToLower(mac.String())
}

func findARPIPByMAC(macAddress string) (string, bool) {
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return "", false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 || fields[0] == "IP" {
			continue
		}

		if strings.EqualFold(fields[3], macAddress) && fields[2] == "0x2" {
			return fields[0], true
		}
	}

	return "", false
}

func probeSubnet(broadcastAddress string, port int) {
	broadcastIP := net.ParseIP(broadcastAddress).To4()
	if broadcastIP == nil {
		return
	}

	targets := probeTargetsForBroadcast(broadcastIP)
	if len(targets) == 0 {
		return
	}

	if port <= 0 {
		port = 9
	}

	var wg sync.WaitGroup
	limit := make(chan struct{}, 64)
	for _, target := range targets {
		wg.Add(1)
		limit <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-limit }()

			conn, err := net.DialTimeout("udp4", net.JoinHostPort(ip, strconv.Itoa(port)), 150*time.Millisecond)
			if err != nil {
				return
			}
			defer conn.Close()
			_, _ = conn.Write([]byte{0})
		}(target)
	}
	wg.Wait()
}

func probeTargetsForBroadcast(broadcastIP net.IP) []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	targets := []string{}
	scanAllLocalSubnets := broadcastIP.Equal(net.IPv4(255, 255, 255, 255))
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}

			ones, bits := ipNet.Mask.Size()
			if bits != 32 || ones < 23 {
				continue
			}

			if !scanAllLocalSubnets && !calculatedBroadcast(ip, ipNet.Mask).Equal(broadcastIP) {
				continue
			}

			if scanAllLocalSubnets {
				targets = append(targets, subnetHosts(ipNet)...)
				continue
			}

			return subnetHosts(ipNet)
		}
	}

	if scanAllLocalSubnets {
		return targets
	}

	if !isLikelyBroadcast(broadcastIP) {
		return []string{broadcastIP.String()}
	}

	return nil
}

func calculatedBroadcast(ip net.IP, mask net.IPMask) net.IP {
	result := make(net.IP, net.IPv4len)
	for i := range result {
		result[i] = ip[i] | ^mask[i]
	}
	return result
}

func subnetHosts(ipNet *net.IPNet) []string {
	ip := ipNet.IP.To4()
	if ip == nil {
		return nil
	}

	ones, bits := ipNet.Mask.Size()
	if bits != 32 || ones < 23 {
		return nil
	}

	start := binary.BigEndian.Uint32(ip.Mask(ipNet.Mask))
	size := uint32(1) << uint32(32-ones)
	if size <= 2 || size > 512 {
		return nil
	}

	hosts := make([]string, 0, size-2)
	for current := start + 1; current < start+size-1; current++ {
		next := make(net.IP, net.IPv4len)
		binary.BigEndian.PutUint32(next, current)
		hosts = append(hosts, next.String())
	}

	return hosts
}

func isLikelyBroadcast(ip net.IP) bool {
	return ip[3] == 255 || ip.Equal(net.IPv4(255, 255, 255, 255))
}
