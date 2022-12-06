package ServerStatus

type ServerStatus struct {
    Name              string              `json:"name"`
    LastSeen          int64               `json:"last_seen"`
    Uptime            uint64              `json:"uptime"`
    Load              [3]float64          `json:"load"`
    Network           []*NetworkInterface `json:"network_interface,omitempty"`
    Disk              []*DiskInfo         `json:"disk,omitempty"`
    Partition         []*Partition        `json:"partition,omitempty"`
    Cpu               string              `json:"cpu"`
    CpuUsedPercent    float64             `json:"cpu_used_percent"`
    MemoryTotal       uint64              `json:"memory_total"`
    MemoryUsedPercent float64             `json:"memory_used_percent"`
    SwapTotal         uint64              `json:"swap_total"`
    SwapUsedPercent   float64             `json:"swap_used_percent"`
    CpuVersion        string              `json:"cpu_version"`
    CpuArch           string              `json:"cpu_arch"`
    TcpNum            int                 `json:"tcp_num"`
    UdpNum            int                 `json:"udp_num"`
}
type NetworkInterface struct {
    Name  string `json:"name"`
    RX    uint64 `json:"rx"`
    TX    uint64 `json:"tx"`
    RXps  uint64 `json:"rx_ps"`
    TXps  uint64 `json:"tx_ps"`
    RXP   uint64 `json:"rxp"`
    TXP   uint64 `json:"txp"`
    RXPps uint64 `json:"rxp_ps"`
    TXPps uint64 `json:"txp_ps"`
}
type DiskInfo struct {
    Name             string `json:"name"`
    Device           string `json:"device"`
    ReadCount        uint64 `json:"r"`
    WriteCount       uint64 `json:"w"`
    ReadBytes        uint64 `json:"rb"`
    WriteBytes       uint64 `json:"wb"`
    ReadCountPerSec  uint64 `json:"rps"`
    WriteCountPerSec uint64 `json:"wps"`
    ReadBytesPerSec  uint64 `json:"rbps"`
    WriteBytesPerSec uint64 `json:"wbps"`
}
type Partition struct {
    Path              string  `json:"path"`
    FSType            string  `json:"fs_type"`
    Total             uint64  `json:"total"`
    Free              uint64  `json:"free"`
    Used              uint64  `json:"used"`
    UsedPercent       float64 `json:"used_percent"`
    InodesTotal       uint64  `json:"inodes_total"`
    InodesUsed        uint64  `json:"inodes_used"`
    InodesFree        uint64  `json:"inodes_free"`
    InodesUsedPercent float64 `json:"inodes_used_percent"`
}
