package libusb

/*
#include <libusb-1.0/libusb.h>

extern void cgo_transfer_cb_fn(struct libusb_transfer* transfer);
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

// Transfer status codes.
const (
	TRANSFER_COMPLETED = C.LIBUSB_TRANSFER_COMPLETED
	TRANSFER_ERROR     = C.LIBUSB_TRANSFER_ERROR
	TRANSFER_TIMED_OUT = C.LIBUSB_TRANSFER_TIMED_OUT
	TRANSFER_CANCELLED = C.LIBUSB_TRANSFER_CANCELLED
	TRANSFER_STALL     = C.LIBUSB_TRANSFER_STALL
	TRANSFER_NO_DEVICE = C.LIBUSB_TRANSFER_NO_DEVICE
	TRANSFER_OVERFLOW  = C.LIBUSB_TRANSFER_OVERFLOW
)

// Transfer.Flags values.
const (
	TRANSFER_SHORT_NOT_OK    = C.LIBUSB_TRANSFER_SHORT_NOT_OK
	TRANSFER_FREE_BUFFER     = C.LIBUSB_TRANSFER_FREE_BUFFER
	TRANSFER_FREE_TRANSFER   = C.LIBUSB_TRANSFER_FREE_TRANSFER
	TRANSFER_ADD_ZERO_PACKET = C.LIBUSB_TRANSFER_ADD_ZERO_PACKET
)

var _stat_msg map[byte]string = map[byte]string{
	TRANSFER_COMPLETED: "TRANSFER_COMPLETED",
	TRANSFER_ERROR:	 	"TRANSFER_ERROR",
	TRANSFER_TIMED_OUT: "TRANSFER_TIMED_OUT",
	TRANSFER_CANCELLED: "TRANSFER_CANCELLED",
	TRANSFER_STALL: 	"TRANSFER_STALL",
	TRANSFER_NO_DEVICE: "TRANSFER_NO_DEVICE",
	TRANSFER_OVERFLOW: 	"TRANSFER_OVERFLOW",
}

// Internal stucture needed to keep track
// of transfers from GO
type t_transfer_map struct {
	sync.Map
}

var _transfer_map t_transfer_map = t_transfer_map{}

func (tm *t_transfer_map)add(tr *C.struct_libusb_transfer, gotr *Transfer) {
	tm.Map.Store(tr,gotr)
}

func (tm *t_transfer_map)del(tr *C.struct_libusb_transfer) {
	tm.Map.Delete(tr)
}

func (tm *t_transfer_map)get(tr *C.struct_libusb_transfer) *Transfer {
	if tmp, ok := tm.Map.Load(tr); ok {
		return tmp.(*Transfer)
	}
	return nil
}

func (tm *t_transfer_map)has(tr *C.struct_libusb_transfer) bool {
	_, ok := tm.Map.Load(tr)
	return ok
}

/*

struct libusb_transfer {
	libusb_device_handle *dev_handle;
	uint8_t flags;
	unsigned char endpoint;
	unsigned char type;
	unsigned int timeout;
	enum libusb_transfer_status status;
	int length;
	int actual_length;
	libusb_transfer_cb_fn callback;
	void *user_data;
	unsigned char *buffer;
	int num_iso_packets;
	struct libusb_iso_packet_descriptor iso_packet_desc[];
};

*/

// The generic USB transfer structure. The user populates this structure and
// then submits it in order to request a transfer. After the transfer has
// completed, the library populates the transfer with the results and passes
// it back to the user.
type Transfer_Cb_Fn func(transfer *Transfer)

type Transfer struct {
	ptr 			*C.struct_libusb_transfer
	handle 			Device_Handle
	callback		func(*Transfer)
	user_data		interface{}
	buffer			*[]byte
}

func (tr *Transfer)GetPtr() *C.struct_libusb_transfer { return tr.ptr }
func (tr *Transfer)GetHandle() Device_Handle          { return tr.handle }
func (tr *Transfer)GetFlags() uint8                   { return uint8(tr.ptr.flags) }
func (tr *Transfer)GetEndpoint() uint8                { return uint8(tr.ptr.endpoint) }
func (tr *Transfer)GetType() uint8                    { return uint8(tr.ptr._type) }
func (tr *Transfer)GetTimeout() int                   { return int(tr.ptr.timeout) }
func (tr *Transfer)GetStatus() uint8                  { return uint8(tr.ptr.status) }
func (tr *Transfer)GetStatusStr() string              { return _stat_msg[tr.GetStatus()] }
func (tr *Transfer)GetLength() int                    { return int(tr.ptr.length) }
func (tr *Transfer)GetActualLength() int              { return int(tr.ptr.actual_length) }
func (tr *Transfer)GetCallback() func(*Transfer)      { return tr.callback }
func (tr *Transfer)GetUserData() interface{}          { return tr.user_data }
func (tr *Transfer)GetBuffer() *[]byte                { return tr.buffer }
func (tr *Transfer)GetData() []byte {
	tmp := make([]byte, tr.GetActualLength())
	copy(tmp,*tr.buffer)
	return tmp
	//return (*tr.buffer)[:tr.GetActualLength()]
}

func (tr *Transfer)SetFlags(flags uint8) {
	tr.ptr.flags = (C.uint8_t)(flags)
}

func (tr *Transfer)SetCallback(cb func(*Transfer)) {
	tr.callback = cb
}

func (tr *Transfer)SetUserData(user_data interface{}) {
	tr.user_data = user_data
}
//
//func c2go_Transfer(x *C.struct_libusb_transfer) *Transfer {
//	transfer := _transfer_map.get(x)
//	if transfer == nil || transfer.ptr != x { return nil }
//	return transfer
//}
//
//func go2c_Transfer(x *Transfer) *C.struct_libusb_transfer {
//	return x.ptr
//}

//-----------------------------------------------------------------------------
// static struct libusb_control_setup * 	libusb_control_transfer_get_setup (struct libusb_transfer *transfer)
// static void 	libusb_fill_control_setup (unsigned char *buffer, uint8_t bmRequestType, uint8_t bRequest, uint16_t wValue, uint16_t wIndex, uint16_t wLength)
// static void 	libusb_fill_control_transfer (struct libusb_transfer *transfer, libusb_device_handle *dev_handle, unsigned char *buffer, libusb_transfer_cb_fn callback, void *user_data, unsigned int timeout)
// static void 	libusb_fill_bulk_transfer (struct libusb_transfer *transfer, libusb_device_handle *dev_handle, unsigned char endpoint, unsigned char *buffer, int length, libusb_transfer_cb_fn callback, void *user_data, unsigned int timeout)
// static void 	libusb_fill_bulk_stream_transfer (struct libusb_transfer *transfer, libusb_device_handle *dev_handle, unsigned char endpoint, uint32_t stream_id, unsigned char *buffer, int length, libusb_transfer_cb_fn callback, void *user_data, unsigned int timeout)
// static void 	libusb_fill_interrupt_transfer (struct libusb_transfer *transfer, libusb_device_handle *dev_handle, unsigned char endpoint, unsigned char *buffer, int length, libusb_transfer_cb_fn callback, void *user_data, unsigned int timeout)
// static void 	libusb_fill_iso_transfer (struct libusb_transfer *transfer, libusb_device_handle *dev_handle, unsigned char endpoint, unsigned char *buffer, int length, int num_iso_packets, libusb_transfer_cb_fn callback, void *user_data, unsigned int timeout)
// static void 	libusb_set_iso_packet_lengths (struct libusb_transfer *transfer, unsigned int length)
// static unsigned char * 	libusb_get_iso_packet_buffer (struct libusb_transfer *transfer, unsigned int packet)
// static unsigned char * 	libusb_get_iso_packet_buffer_simple (struct libusb_transfer *transfer, unsigned int packet)

//-----------------------------------------------------------------------------
//Asynchronous device I/O

func Alloc_Streams(dev Device_Handle, num_streams uint32, endpoints []byte) (int, error) {
	rc := int(C.libusb_alloc_streams(dev, (C.uint32_t)(num_streams), (*C.uchar)(&endpoints[0]), (C.int)(len(endpoints))))
	if rc < 0 {
		return 0, &libusb_error{rc}
	}
	return rc, nil
}

func Free_Streams(dev Device_Handle, endpoints []byte) error {
	rc := int(C.libusb_free_streams(dev, (*C.uchar)(&endpoints[0]), (C.int)(len(endpoints))))
	if rc != 0 {
		return &libusb_error{rc}
	}
	return nil
}

func Transfer_Set_Stream_ID(transfer *Transfer, stream_id uint32) {
	C.libusb_transfer_set_stream_id(transfer.ptr, (C.uint32_t)(stream_id))
}

func Transfer_Get_Stream_ID(transfer *Transfer) uint32 {
	return uint32(C.libusb_transfer_get_stream_id(transfer.ptr))
}

func Control_Transfer_Get_Data(transfer *Transfer) *byte {
  // should this return a slice? - what's the length?
	return (*byte)(C.libusb_control_transfer_get_data(transfer.ptr))
}

func Alloc_Transfer(iso_packets int) (*Transfer, error) {
	ptr := C.libusb_alloc_transfer((C.int)(iso_packets))
	if ptr == nil {
		return nil, &libusb_error{ERROR_OTHER}
	}
	transfer := new(Transfer)
	transfer.ptr = ptr

	_transfer_map.add(ptr, transfer)
	return transfer, nil
}

func (tr *Transfer)Free() {
	_transfer_map.del(tr.ptr)
	C.libusb_free_transfer(tr.ptr)
}

func (tr *Transfer)Submit() error {
	rc := int(C.libusb_submit_transfer(tr.ptr))
	if rc != 0 {
		return &libusb_error{rc}
	}
	return nil
}

func (tr *Transfer)Cancel() error {
	_transfer_map.del(tr.ptr)
	rc := int(C.libusb_cancel_transfer(tr.ptr))
	if rc != 0 {
		return &libusb_error{rc}
	}
	return nil
}


//export transfer_cb_fn
func transfer_cb_fn(tr *C.struct_libusb_transfer) {
	transfer := _transfer_map.get(tr)
	if transfer == nil || transfer.ptr != tr {
		// Looks like a transfer was canceled in progress.
		// NOTE: It's your responsibility to ensure that the transfer
		// struct was freed properly.
		return
	}
	if transfer.callback != nil {
		transfer.callback(transfer)
	}
}

func (tr *Transfer)Fill_Bulk(hdl Device_Handle, endpoint uint8,
	data []byte, timeout uint) error {

	// At this point the transfer object must already be in the transfer map
	if !_transfer_map.has(tr.ptr) {
		return fmt.Errorf("Transfer not allocated. Please call Alloc_Transfer first.")
	}

	tr.handle = hdl
	tr.buffer = &data

	C.libusb_fill_bulk_transfer(
		tr.ptr, hdl, C.uchar(endpoint), (*C.uchar)(&data[0]), C.int(len(data)),
		C.libusb_transfer_cb_fn(unsafe.Pointer(C.cgo_transfer_cb_fn)),
		nil, C.uint(timeout),
	)
	return nil
}
