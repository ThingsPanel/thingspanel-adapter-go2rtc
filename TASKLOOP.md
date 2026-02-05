## TaskLoop Run - ThingsPanel go2rtc End-to-End

### Requirements (from user)
- Use taskloop skill to test and fully connect the pipeline.
- Goal: Mac camera -> FFmpeg -> go2rtc -> ThingsPanel device sync.

### Assumptions
- I do not have direct access to the private LAN or servers.
- User will run provided commands on Mac and the go2rtc/adapter host.

### Acceptance Criteria
- go2rtc API shows the pushed stream (e.g., `mac_cam`).
- Adapter device list returns `list: []` (not `null`) when empty.
- ThingsPanel sync shows the stream device.

### References (top 3)
- go2rtc RTSP listen config default 8554.
- ffmpeg RTSP `listen` flag behavior.
- go2rtc project docs (default ports).

### Plan Split
1) Fix adapter `list:null` bug by redeploying v1.0.1+ binary.
2) Validate go2rtc RTSP publish or pull workflow.
3) Verify ThingsPanel sync shows device.

### Iterations
#### Iteration 1 (2026-02-05)
- Found adapter responding `list:null` => indicates old binary.
- Proposed recompile+deploy and retest.
- Determined RTSP publish connection refused; need go2rtc RTSP config or auth.

### Status
- Pending user-run commands to proceed.
