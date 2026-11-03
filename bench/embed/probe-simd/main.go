package main
import ("bufio";"encoding/binary";"encoding/json";"fmt";"os";"runtime";"runtime/debug";"runtime/pprof";"strconv";"time";"github.com/rcarmo/gte-go/gte")
func main(){ debug.SetGCPercent(100); t0:=time.Now(); m,err:=gte.Load(os.Args[1]); if err!=nil{panic(err)}; load:=time.Since(t0)
 f,_:=os.Open(os.Args[2]); s:=bufio.NewScanner(f); s.Buffer(make([]byte,1<<24),1<<24); var ts []string; for s.Scan(){var r struct{Text string `json:"text"`}; json.Unmarshal(s.Bytes(),&r); ts=append(ts,r.Text)}
 w,_:=strconv.Atoi(os.Args[4]); out,_:=os.Create(os.Args[3]); defer out.Close(); if p:=os.Getenv("CPUPROF");p!=""{f,_:=os.Create(p);pprof.StartCPUProfile(f);defer pprof.StopCPUProfile()}; t1:=time.Now(); var v [][]float32
 if w<=1 { for _,x:=range ts { e,err:=m.Embed(x); if err!=nil{panic(err)}; v=append(v,append([]float32(nil),e...)) } } else { v,err=m.EmbedBatchParallel(ts,w); if err!=nil{panic(err)} }
 d:=time.Since(t1); bw:=bufio.NewWriter(out); for _,e:=range v { binary.Write(bw,binary.LittleEndian,e) }; bw.Flush()
 var ms runtime.MemStats; runtime.ReadMemStats(&ms); fmt.Printf("load=%.2fs n=%d workers=%d secs=%.2f per_s=%.1f ms_each=%.1f dim=%d sys_mb=%d maxlen=%d\n",load.Seconds(),len(ts),w,d.Seconds(),float64(len(ts))/d.Seconds(),d.Seconds()*1000/float64(len(ts)),len(v[0]),ms.Sys>>20,m.MaxLen())}
