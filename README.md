This is a fork of deadsys libusb library wrpaper.(https://godoc.org/github.com/deadsy/libusb)

# libusb
golang wrapper for libusb-1.0

The API for libusb has been mapped 1-1 to equivalent go functions and types
except for the async functions.  Due to how GOC interacts with native libraries
an additional layer had to be added to deal with callbacks and buffer handling.

See http://libusb.info/ for more information on the C-API

I've added initial async functions for bulk transfers.

Example:

Somewhere in your code add a thread that does the listening:

```go
func usb_listen(done chan struct{}) {

	// Loop forever to keep reading
	// We add a 500ms timeout to handle other events if needed
	var tv syscall.Timeval = syscall.Timeval{0,500*1000}
	for {
		select {
		case <- done:
			break
		default:
			libusb.Handle_Events_Timeout_Completed(nil,tv)
		}
	}
}
```

Then add the functions to pull data from the device:
```go
func (dev *UBSObj)read_callback(transfer *libusb.Transfer) {
    data := transfer.GetData()

	if transfer.GetStatus() == libusb.TRANSFER_COMPLETED {
		// Do it again
		err := transfer.Submit()
		if err != nil {
			// ERROR CONDITION
		}
	} else {
		// We received a status other than completed.
		// Return back to the main loop and singal end of stream
		// Most likely the USB device was pulled.
		// Cancel any pending transaction and free it
		transfer.Cancel()
		transfer.Free()
	}
}

func (dev *USBObj)read_usb() error {
    transfer, err := libusb.Alloc_Transfer(0)
    if err != nil {
        return err
    }

    // Setup the receive buffer
    buf := make([]byte,2048)

    // Configure the transfer and attach the buffer
    transfer.Fill_Bulk(dev.hdev,dev.ep_read,buf,0)

    // Set the GO callback function
    transfer.SetCallback(dev.read_callback)

    // Submit the initial read transfer
    err = transfer.Submit()
    if err != nil {
        return err
    }
    return nil
}
```

NOTE: you can call read_usb everytime you want to read, or you can have the
callback autoloop for you, resubmitting the transfer on each packet received.

For writes just change the ep_read value to the write value.

