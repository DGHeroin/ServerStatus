package agent

import (
    "bytes"
    "encoding/json"
    "fmt"
    . "github.com/DGHeroin/ServerStatus/ServerStatus"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/disk"
    "github.com/shirou/gopsutil/v3/host"
    "github.com/shirou/gopsutil/v3/load"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/net"
    "github.com/spf13/cobra"
    "io/ioutil"
    "math"
    "net/http"
    "os"
    "runtime"
    "strconv"
    "time"
)

var (
    Cmd = &cobra.Command{
        Use: "agent args",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runAgent()
        },
    }
)

var (
    sid            string
    addr           string
    auth           string
    isDebug        bool
    netDevs        []string
    diskDevs       []string
    partitionsDevs []string
)

func init() {
    Cmd.PersistentFlags().StringVar(&sid, "s", "🐱", "node name")
    Cmd.PersistentFlags().StringVar(&addr, "addr", "http://127.0.0.1:8080", "server base url")
    Cmd.PersistentFlags().StringVar(&auth, "auth", "", "agent auth")
    Cmd.PersistentFlags().BoolVar(&isDebug, "debug", false, "is debug")
    Cmd.PersistentFlags().StringSliceVar(&netDevs, "net", []string{}, "network interface")
    Cmd.PersistentFlags().StringSliceVar(&diskDevs, "disk", []string{}, "disks")
    Cmd.PersistentFlags().StringSliceVar(&partitionsDevs, "partitions", []string{}, "partitions")

}
func monitor(ch chan *ServerStatus) {
    lastNetwork := getNetworkInterfaces(nil)
    lastDisk := getDisk(nil)
    for {
        time.Sleep(1 * time.Second)
        go func() {
            v, _ := mem.VirtualMemory()
            info, _ := cpu.Info()
            c, _ := cpu.Percent(0*time.Second, false)
            uptime, _ := host.Uptime()
            arch, _ := host.KernelArch()
            version, _ := host.KernelVersion()
            tcpNum, _ := net.Connections("tcp")
            udpNum, _ := net.Connections("udp")
            swap, _ := mem.SwapMemory()
            networks := getNetworkInterfaces(lastNetwork)
            disks := getDisk(lastDisk)
            partition := getPartition()
            var loads = [3]float64{}
            if avg, err := load.Avg(); err == nil {
                loads = [3]float64{avg.Load1, avg.Load5, avg.Load15}
            }
            server := &ServerStatus{
                Name:              sid,
                Uptime:            uptime,
                Load:              loads,
                Network:           networks,
                Disk:              disks,
                Partition:         partition,
                Cpu:               strconv.Itoa(runtime.NumCPU()) + "*" + info[0].ModelName,
                CpuUsedPercent:    math.Round(c[0]*1000) / 1000,
                CpuVersion:        version,
                CpuArch:           arch,
                MemoryTotal:       v.Total,
                MemoryUsedPercent: math.Round(v.UsedPercent*1000) / 1000,
                SwapTotal:         swap.Total,
                SwapUsedPercent:   math.Round(swap.UsedPercent*1000) / 1000,
                TcpNum:            len(tcpNum),
                UdpNum:            len(udpNum),
            }

            lastNetwork = networks
            ch <- server
        }()
    }
}
func runAgent() error {
    ch := make(chan *ServerStatus, 1000)
    go func() {
        for status := range ch {
            str, _ := json.Marshal(status)
            if isDebug {
                _, _ = fmt.Fprintln(os.Stdout, string(str))
            }
            if addr == "" {
                continue
            }
            request, err := http.NewRequest(http.MethodPost, addr+"/api/agent/_post", bytes.NewBuffer(str))
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                continue
            }
            request.Header.Add("Content-Type", "application/json")
            request.Header.Add("Authorization", auth)
            resp, err := http.DefaultClient.Do(request)
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                continue
            }
            if resp.StatusCode != http.StatusOK {
                data, err := ioutil.ReadAll(resp.Body)
                _, _ = fmt.Fprintln(os.Stderr, err, string(data))
            }
        }
    }()

    monitor(ch)
    return nil
}

func getNetworkInterfaces(last []*NetworkInterface) (interfaces []*NetworkInterface) {
    counters, _ := net.IOCounters(true)
    for i := range counters {
        itf := counters[i]
        if !containsStringsZeroTrue(netDevs, itf.Name) {
            continue
        }
        interfaces = append(interfaces, &NetworkInterface{
            Name:     itf.Name,
            RX:       itf.BytesRecv,
            TX:       itf.BytesSent,
            RXPacket: itf.PacketsRecv,
            TXPacket: itf.PacketsSent,
        })

    }
    for _, d := range interfaces {
        for _, n := range last {
            if n.Name == d.Name {
                d.RXPerSec = d.RX - n.RX
                d.TXPerSec = d.TX - n.TX

                d.RXPacketPerSec = d.RXPacket - n.RXPacket
                d.TXPacketPerSec = d.TXPacket - n.TXPacket
            }
        }
    }
    return
}
func getDisk(last []*DiskInfo) (disks []*DiskInfo) {
    counters, _ := disk.IOCounters()

    for i := range counters {
        itf := counters[i]
        if !containsStringsZeroTrue(diskDevs, itf.Name) {
            continue
        }
        disks = append(disks, &DiskInfo{
            Name:       itf.Name,
            ReadCount:  itf.ReadCount,
            WriteCount: itf.WriteCount,
            ReadBytes:  itf.ReadBytes,
            WriteBytes: itf.WriteBytes,
        })
    }
    for _, d := range disks {
        for _, n := range last {
            if n.Name == d.Name {
                d.ReadCountPerSec = d.ReadCount - n.ReadCount
                d.WriteCountPerSec = d.WriteCount - n.WriteCount

                d.ReadBytesPerSec = d.ReadBytes - n.ReadBytes
                d.WriteBytesPerSec = d.WriteBytes - n.WriteBytes
            }
        }
    }
    return
}
func getPartition() (result []*Partition) {
    partitions, _ := disk.Partitions(true)
    for _, info := range partitions {
        usage, err := disk.Usage(info.Mountpoint)
        if err != nil {
            continue
        }
        if !containsStringsZeroTrue(partitionsDevs, info.Mountpoint) {
            continue
        }
        result = append(result, &Partition{
            Path:              usage.Path,
            FSType:            usage.Fstype,
            Total:             usage.Total,
            Free:              usage.Free,
            Used:              usage.Used,
            UsedPercent:       usage.UsedPercent,
            InodesTotal:       usage.InodesTotal,
            InodesUsed:        usage.InodesUsed,
            InodesFree:        usage.InodesFree,
            InodesUsedPercent: usage.InodesUsedPercent,
        })
    }
    return
}

func containsStringsZeroTrue(ss []string, s string) bool {
    if len(ss) == 0 {
        return true
    }
    for _, v := range ss {
        if v == s {
            return true
        }
    }
    return false
}

func containsStringsZeroFalse(ss []string, s string) bool {
    if len(ss) == 0 {
        return false
    }
    for _, v := range ss {
        if v == s {
            return true
        }
    }
    return false
}
