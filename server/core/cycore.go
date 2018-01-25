package core
// Environment: LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/bahadir/cuda-workspace/cycoregpu/Debug:/usr/local/cuda-9.1/lib64

//#cgo CFLAGS: -I../../core
//#cgo LDFLAGS: -L../../core/Release -lxenon-core
//#include <CyclopsCore.h>
import "C"
import "unsafe"

func initializeStorage(maxSize uint64) {
	C.initStorage(C.uint64_t(maxSize))
}

func addHash(hash uint64) {
	C.addHash(C.uint64_t(hash))
}

func search(hash uint64, distance uint64) []uint64 {

	r := C.search(C.uint64_t(hash), C.uint64_t(distance))
	slice := (*[1 << 30] C.uint64_t)(unsafe.Pointer(r.hashPtr))[:r.size:r.size]
	result := make([]uint64, uint64(r.size) -1)

	for i := uint64(1); i < uint64(r.size); i++{
		result[i-1] = uint64(slice[i])
	}

	return result
}



