package libusb

/*
#include <libusb-1.0/libusb.h>

extern void cgo_transfer_cb_fn(struct libusb_transfer* transfer);
*/
import "C"

import "fmt"
import "sync"
import "unsafe"

// Internal stucture needed to keep track
// of transfers from GO
type t_transfer_map struct {
	sync.Mutex
	_map map[*C.struct_libusb_transfer]*Transfer
}

var _transfer_map t_transfer_map = t_transfer_map{
	_map:  map[*C.struct_libusb_transfer]*Transfer{},
}

func (self *t_transfer_map)add(c_tr *C.struct_libusb_transfer, gotr *Transfer) {
	self.Lock()
	defer self.Unlock()
	self._map[c_tr] = gotr
}

func (self *t_transfer_map)del(tr *C.struct_libusb_transfer) {
	self.Lock()
	defer self.Unlock()
	delete(self._map,tr)
}

func (self *t_transfer_map)get(tr *C.struct_libusb_transfer) *Transfer {
	self.Lock()
	defer self.Unlock()
	return self._map[tr]
}

func (self *t_transfer_map)has(tr *C.struct_libusb_transfer) bool {
	self.Lock()
	defer self.Unlock()
	_, ok := self._map[tr]
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

func (self *Transfer)GetHandle() Device_Handle { return self.handle }
func (self *Transfer)GetFlags() uint8 { return uint8(self.ptr.flags) }
func (self *Transfer)GetEndpoint() uint8 { return uint8(self.ptr.endpoint) }
func (self *Transfer)GetType() uint8 { return uint8(self.ptr._type) }
func (self *Transfer)GetTimeout() int { return int(self.ptr.timeout) }
func (self *Transfer)GetStatus() uint8 { return uint8(self.ptr.status) }
func (self *Transfer)GetLength() int { return int(self.ptr.length) }
func (self *Transfer)GetActualLength() int { return int(self.ptr.actual_length) }
func (self *Transfer)GetCallback() func(*Transfer) { return self.callback }
func (self *Transfer)GetUserData() interface{} { return self.user_data }
func (self *Transfer)GetBuffer() *[]byte { return self.buffer }
func (self *Transfer)GetData() []byte {
	tmp := make([]byte,self.GetActualLength())
	copy(tmp,*self.buffer)
	return tmp
	//return (*self.buffer)[:self.GetActualLength()]
}

func (self *Transfer)SetCallback(cb func(*Transfer)) {
	self.callback = cb
}

func (self *Transfer)SetUserData(user_data interface{}) {
	self.user_data = user_data
}

func c2go_Transfer(x *C.struct_libusb_transfer) *Transfer {
	transfer := _transfer_map.get(x)
	if transfer.ptr != x { panic("no ptr match") }
	return transfer
}

func go2c_Transfer(x *Transfer) *C.struct_libusb_transfer {
	return x.ptr
}

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
	C.libusb_transfer_set_stream_id(go2c_Transfer(transfer), (C.uint32_t)(stream_id))
}

func Transfer_Get_Stream_ID(transfer *Transfer) uint32 {
	return uint32(C.libusb_transfer_get_stream_id(go2c_Transfer(transfer)))
}

func Control_Transfer_Get_Data(transfer *Transfer) *byte {
  // should this return a slice? - what's the length?
	return (*byte)(C.libusb_control_transfer_get_data(go2c_Transfer(transfer)))
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

func Free_Transfer(transfer *Transfer) {
	_transfer_map.del(transfer.ptr)
	C.libusb_free_transfer(transfer.ptr)
}

func Submit_Transfer(transfer *Transfer) error {
	rc := int(C.libusb_submit_transfer(go2c_Transfer(transfer)))
	if rc != 0 {
		return &libusb_error{rc}
	}
	return nil
}

func Cancel_Transfer(transfer *Transfer) error {
	rc := int(C.libusb_cancel_transfer(go2c_Transfer(transfer)))
	if rc != 0 {
		return &libusb_error{rc}
	}
	return nil
}


//export transfer_cb_fn
func transfer_cb_fn(tr *C.struct_libusb_transfer) {
	transfer := c2go_Transfer(tr)
	transfer.callback(transfer)
}

func (self *Transfer)Fill_Bulk(hdl Device_Handle, endpoint uint8,
	data []byte, timeout uint) error {

	// At this point the transfer object must already be in the transfer map
	if !_transfer_map.has(self.ptr) {
		return fmt.Errorf("Transfer not allocated. Please call Alloc_Transfer first.")
	}

	self.handle = hdl
	self.buffer = &data

	C.libusb_fill_bulk_transfer(
		self.ptr, hdl, C.uchar(endpoint), (*C.uchar)(&data[0]), C.int(len(data)),
		C.libusb_transfer_cb_fn(unsafe.Pointer(C.cgo_transfer_cb_fn)),
		nil, C.uint(timeout),
	)
	return nil
}
