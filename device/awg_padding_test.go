/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2025 WireGuard LLC. All Rights Reserved.
 */

package device

import (
	"testing"
)

// The TUN reader waits for a packet with the offset of the S4 it found before
// the wait. A packet that arrives after S4 was set must go out with the new S4,
// or the receiver looks for the transport header at another offset and drops
// it: the first packet after every change of S4, and the first packet of a
// device that got its S4 after it was created.
func TestTransportPaddingSetWhileReaderWaits(t *testing.T) {
	goroutineLeakCheck(t)

	pair := genTestPair(t, false)

	setPadding := func(value string) {
		t.Helper()
		for i := range pair {
			if err := pair[i].dev.IpcSet(uapiCfg("s4", value)); err != nil {
				t.Fatalf("failed to set s4=%s on device %d: %v", value, i, err)
			}
		}
	}
	ping := func(name string) {
		t.Helper()
		t.Run(name+" ping 1.0.0.1", func(t *testing.T) {
			pair.Send(t, Ping, nil)
		})
		t.Run(name+" ping 1.0.0.2", func(t *testing.T) {
			pair.Send(t, Pong, nil)
		})
	}

	// After these packets both readers wait again, with the offset of S4 = 0.
	ping("without padding")

	// The packets move towards the end of their buffers.
	setPadding("25")
	ping("padding added")

	// And back towards the start.
	setPadding("7")
	ping("padding reduced")

	setPadding("0")
	ping("padding removed")
}
