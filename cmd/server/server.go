package server

import (
    "bytes"
    "encoding/json"
    "fmt"
    . "github.com/DGHeroin/ServerStatus/ServerStatus"
    "github.com/dustin/go-humanize"
    "github.com/gin-gonic/gin"
    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cobra"
    "net/http"
    "os"
    "sort"
    "sync"
    "time"
)

var (
    Cmd = &cobra.Command{
        Use: "server <args>",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runServer()
        },
    }
)

var (
    addr      string
    authAgent string
    authView  string
    isDebug   bool
)

func init() {
    Cmd.PersistentFlags()
    Cmd.PersistentFlags().StringVar(&addr, "addr", ":8080", "http listen address")
    Cmd.PersistentFlags().StringVar(&authAgent, "authAgent", RandStringRunes(64), "agent token")
    Cmd.PersistentFlags().StringVar(&authView, "authView", RandStringRunes(64), "http api token")
    Cmd.PersistentFlags().BoolVar(&isDebug, "debug", false, "is debug")

}

func runServer() error {
    if isDebug {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }

    fmt.Printf("agent auth: %s\n", authAgent)
    fmt.Printf("http auth: %s\n", authView)

    var (
        mu sync.RWMutex
        m  = map[string]*ServerStatus{}
    )

    r := gin.New()
    r.GET("api/agent/_flush", doAuth(authAgent), func(c *gin.Context) {
        mu.Lock()
        defer mu.Unlock()
        m = map[string]*ServerStatus{}
    })
    r.GET("api/agent/_kick", doAuth(authAgent), func(c *gin.Context) {
        name := c.Query("name")
        mu.Lock()
        defer mu.Unlock()
        delete(m, name)
    })
    r.POST("api/agent/_post", doAuth(authAgent), func(c *gin.Context) {
        status := &ServerStatus{}
        if err := c.ShouldBindJSON(status); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": -2, "msg": err.Error()})
            return
        }
        if isDebug {
            str, _ := json.Marshal(status)
            _, _ = fmt.Fprintln(os.Stdout, string(str))
        }
        c.JSON(http.StatusOK, gin.H{"code": 0})

        mu.Lock()
        defer mu.Unlock()
        // status.LastSeen = time.Now().Unix()
        m[status.Name] = status
    })
    r.GET("api/view/status", doAuth(authView), func(c *gin.Context) {
        mu.RLock()
        defer mu.RUnlock()
        var ss []*ServerStatus
        for _, status := range m {
            ss = append(ss, status)
        }
        if c.Query("txt") == "1" {
            buf := &bytes.Buffer{}
            table := tablewriter.NewWriter(buf)
            {
                sort.SliceStable(ss, func(i, j int) bool {
                    return ss[i].Name < ss[j].Name
                })
                table.SetHeader([]string{"Name", "Uptime", "Last Seen", "Load", "CPU", "Memory", "Partition", "Disk", "IOPS", "TPC/UDP", "Net", "Net Total"})

                for _, s := range ss {
                    // network
                    sumTx := uint64(0)
                    sumRx := uint64(0)
                    sumTxTotal := uint64(0)
                    sumRxTotal := uint64(0)
                    for _, ni := range s.Network {
                        sumTx += ni.TXPerSec
                        sumRx += ni.RXPerSec
                        sumTxTotal += ni.TX
                        sumRxTotal += ni.RX
                    }
                    networkInfo := fmt.Sprintf("↑%v↓%v", humanize.IBytes(sumTx), humanize.IBytes(sumRx))

                    // disk
                    sumDiskR := uint64(0)
                    sumDiskW := uint64(0)
                    sumDiskRps := uint64(0)
                    sumDiskWps := uint64(0)
                    for _, info := range s.Disk {
                        sumDiskR += info.ReadBytesPerSec
                        sumDiskW += info.WriteBytesPerSec

                        sumDiskRps += info.ReadCountPerSec
                        sumDiskWps += info.WriteCountPerSec
                    }

                    diskIOInfo := fmt.Sprintf("↑%v↓%v",
                        humanize.Bytes(sumDiskR), humanize.Bytes(sumDiskR))
                    // partitions
                    sumPartition := uint64(0)
                    sumPartitionFree := uint64(0)
                    for _, partition := range s.Partition {
                        sumPartition += partition.Total
                        sumPartitionFree += partition.Free
                    }
                    partitionInfo := fmt.Sprintf("%.2f%%",
                        float64(sumPartitionFree)/float64(sumPartition)*100.0,
                    )
                    table.Append([]string{
                        s.Name,
                        humanDuration(time.Second * time.Duration(s.Uptime)),
                        time.Now().Sub(time.Unix(s.LastSeen, 0)).Round(time.Millisecond).String(),
                        fmt.Sprintf("%.2f %.2f %.2f", s.Load[0], s.Load[1], s.Load[2]),
                        fmt.Sprintf(`%.2f%%`, s.CpuUsedPercent),
                        fmt.Sprintf(`%.2f%%`, s.MemoryUsedPercent),
                        partitionInfo,
                        diskIOInfo,
                        fmt.Sprintf(`%v/%v`, sumDiskRps, sumDiskWps),
                        fmt.Sprintf(`%v/%v`, s.TcpNum, s.UdpNum),
                        networkInfo,
                        fmt.Sprintf(`%v/%v`, humanize.Bytes(sumTxTotal), humanize.Bytes(sumRxTotal)),
                    })
                }
            }
            table.Render()
            c.Data(http.StatusOK, "text/plain; charset=utf-8", buf.Bytes())
        } else {
            c.JSON(http.StatusOK, gin.H{"code": 0, "data": ss})
        }
    })
    return r.Run(addr)
}
func doAuth(auth string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authStr := c.Request.Header.Get("Authorization")
        if authStr == "" {
            authStr = c.Query("auth")
        }
        if auth != authStr {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "unauthorized"})
            return
        }
        c.Next()
    }
}

func humanDuration(duration time.Duration) string {
    const minute = 60 // 60 秒
    const hour = 3600 // 3600 秒
    const day = hour * 24
    const week = day * 7
    const month = day * 30
    const year = month * 12

    var diff = int64(duration / time.Second) // 给的日期与当前时间戳的差值

    var str string
    if diff > day {
        str = fmt.Sprintf("%.0fd", float64(diff/day))
    } else if diff > hour {
        str = fmt.Sprintf("%.0fh", float64(diff/hour))
    } else if diff > minute {
        str = fmt.Sprintf("%.0fm", float64(diff/minute))
    } else if diff > 1000 {
        str = fmt.Sprintf("%.0fs", float64(diff/1000))
    } else {
        str = "now"
    }

    return str
}
