# unsafehttp

Pessimisticly optimistic `net/http` drop in replacement. 

## Supports

- [ ] websockets
- [x] http/1
- [ ] http/2
- [ ] http/3

## Benchmarks

```bash
goos: linux
goarch: amd64
pkg: github.com/trevatk/unsafehttp
cpu: 13th Gen Intel(R) Core(TM) i5-1340P
BenchmarkServerAllocationsGetUnsafeHTTP-16        293428            120510 ns/op           27185 B/op         180 allocs/op
BenchmarkServerAllocationsGetHTTP-16              292573            125928 ns/op           27359 B/op         181 allocs/op
BenchmarkServerAllocationsPostUnsafeHTTP-16       184378            196428 ns/op           51500 B/op         217 allocs/op
BenchmarkServerAllocationsPostHTTP-16             179594            267884 ns/op           78611 B/op         435 allocs/op
PASS
```