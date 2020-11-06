# Performance test
## Todo
1. (Optional) `./increase_contrack_table.sh`
2. `cd ./server && go run server.go`
3. (Optional) `./topserver.sh` (to watch the  CPU and memory usage).
4. `cd ../client && ./setup.sh $CONNS $WORKERS` <br/>
For example: `./setup.sh 10000 10` creates 100k concurrent connections to the server.
5. (Optional) Profilling: `cd ../server && ./pprof_goroutine.sh`
6. (Optional) Profilling: `./pprof_heap.sh`
7. (Optional) `cd ../client && ./destroy.sh`
<br/>
Sit back and watch how many concurrent connections your machine can handle :sunglasses: