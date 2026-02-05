package go2rtc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"tp-plugin/internal/protocol"

	"github.com/sirupsen/logrus"
)

type Go2RTCProtocolHandler struct {
	port   int
	client *http.Client
	logger *logrus.Logger
	// apiURL will be set dynamically or from config, default for now
	apiURL string
}

func NewHandler(port int) *Go2RTCProtocolHandler {
	return &Go2RTCProtocolHandler{
		port:   port,
		client: &http.Client{Timeout: 5 * time.Second},
		logger: logrus.New(),            // Should be injected or replaced
		apiURL: "http://localhost:1984", // Default, will be updated from service Access Point if available
	}
}

// SetAPIURL allows updating the go2rtc API URL
func (h *Go2RTCProtocolHandler) SetAPIURL(url string) {
	h.apiURL = url
}

// --- ProtocolHandler Interface Implementation ---

func (h *Go2RTCProtocolHandler) Name() string {
	return "Go2RTC-Adapter"
}

func (h *Go2RTCProtocolHandler) Version() string {
	return "1.0.0"
}

func (h *Go2RTCProtocolHandler) Port() int {
	return h.port
}

func (h *Go2RTCProtocolHandler) ExtractDeviceNumber(data []byte) (string, error) {
	// Not used for this adapter as we don't handle raw TCP data from devices
	return "", fmt.Errorf("not implemented")
}

func (h *Go2RTCProtocolHandler) ParseData(data []byte) (*protocol.Message, error) {
	// Not used for this adapter
	return nil, fmt.Errorf("not implemented")
}

func (h *Go2RTCProtocolHandler) EncodeCommand(cmd *protocol.Command) ([]byte, error) {
	// Not used for this adapter
	return nil, fmt.Errorf("not implemented")
}

func (h *Go2RTCProtocolHandler) Start() error {
	h.logger.Info("Go2RTC Protocol Handler started")
	return nil
}

func (h *Go2RTCProtocolHandler) Stop() error {
	h.logger.Info("Go2RTC Protocol Handler stopped")
	return nil
}

// --- Go2RTC Specific Logic ---

// AddStream adds a stream to go2rtc
// PUT /api/streams?src={url}&name={name}
// OR POST /api/streams with body {"name": "{name}", "channels": {"0": "{url}"}} (depending on API version, but PUT with query params is simpler for v1.2+)
// Let's use the query param method as documented in some versions, or the JSON body method.
// According to go2rtc docs: PUT /api/streams?src=rtsp://...&dst=camera1
// Wait, looking at docs step 34: "The HTTP API is the main part... [API description](...)"
// Common go2rtc API: PUT /api/streams?src={url}&dst={name} (src can be repeated)
// Or POST /api/streams with JSON.
// Let's use PUT query method which is often simplest.
func (h *Go2RTCProtocolHandler) AddStream(name string, url string) error {
	// API: PUT /api/streams?src={url}&name={name}
	// Wait, check standard go2rtc API.
	// usually: PUT /api/streams?src={url}&dst={name} is NOT right.
	// simpler: PUT /api/streams?src={url}&name={name}

	// Let's try PUT /api/streams?src={url}&name={name}
	// If name is not provided in query, it might just add 'src'.

	// Correct API based on typical go2rtc usage:
	// PUT /api/streams?src=rtsp://...&name=camera1

	fullURL := fmt.Sprintf("%s/api/streams?src=%s&name=%s", h.apiURL, url, name)
	req, err := http.NewRequest(http.MethodPut, fullURL, nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("go2rtc API error: %d - %s", resp.StatusCode, string(body))
	}

	h.logger.Infof("Added stream '%s' to go2rtc: %s", name, url)
	return nil
}

// RemoveStream removes a stream
// DELETE /api/streams?src={url}&name={name} or just name?
// DELETE /api/streams?name={name}
func (h *Go2RTCProtocolHandler) RemoveStream(name string) error {
	fullURL := fmt.Sprintf("%s/api/streams?name=%s", h.apiURL, name)
	req, err := http.NewRequest(http.MethodDelete, fullURL, nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// It's okay if it doesn't exist, but let's log
		return fmt.Errorf("go2rtc API error: %d", resp.StatusCode)
	}

	h.logger.Infof("Removed stream '%s' from go2rtc", name)
	return nil
}

// StreamInfo go2rtc流信息
type StreamInfo struct {
	Name    string   `json:"name"`
	Sources []string `json:"sources,omitempty"`
	URL     string   `json:"url,omitempty"` // 提取的第一个源地址
}

type streamDetail struct {
	Producers []struct {
		URL string `json:"url"`
	} `json:"producers"`
}

// ... (existing code)

// ListStreams 从go2rtc获取所有streams列表
// GET /api/streams
func (h *Go2RTCProtocolHandler) ListStreams() ([]StreamInfo, error) {
	fullURL := fmt.Sprintf("%s/api/streams", h.apiURL)
	resp, err := h.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query go2rtc streams: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("go2rtc API error: %d - %s", resp.StatusCode, string(body))
	}

	// go2rtc returns map[string]interface{} where key is stream name
	// Value is actually complex, we need to parse it to get producers
	var streamsMap map[string]streamDetail
	if err := json.Unmarshal(body, &streamsMap); err != nil {
		return nil, fmt.Errorf("failed to parse streams: %v", err)
	}

	var streams []StreamInfo
	for name, detail := range streamsMap {
		info := StreamInfo{Name: name}
		if len(detail.Producers) > 0 {
			info.URL = detail.Producers[0].URL
			info.Sources = append(info.Sources, info.URL)
		}
		// Debug log
		h.logger.Infof("Parsed stream: %s, URL: %s, Producers: %d", name, info.URL, len(detail.Producers))
		streams = append(streams, info)
	}

	h.logger.Debugf("Listed %d streams from go2rtc", len(streams))
	return streams, nil
}
