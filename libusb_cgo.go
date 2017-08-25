package libusb

/*
#include <libusb-1.0/libusb.h>

extern void transfer_cb_fn(struct libusb_transfer*);

void cgo_transfer_cb_fn(struct libusb_transfer* transfer) {
	transfer_cb_fn(transfer);
}

*/
import "C"