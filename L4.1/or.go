package or

func Or(channels ...<-chan any) <-chan any {
	switch len(channels) {
	case 0:
		// No channels: return an already-closed channel
		c := make(chan any)
		close(c)
		return c
	case 1:
		// Single channel: return it directly (no need for extra goroutine)
		return channels[0]
	}

	// Multiple channels: spawn a goroutine to wait for any channel to close
	orDone := make(chan any)
	go func() {
		defer close(orDone)

		switch len(channels) {
		case 2:
			// Base case: two channels, use select
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			// Recursive case: divide channels in half and recurse
			m := len(channels) / 2
			select {
			case <-Or(channels[:m]...):
			case <-Or(channels[m:]...):
			}
		}
	}()

	return orDone
}
