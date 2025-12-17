package vna

func (v *VNA) ctxStopped() bool {
	select {
	case <-v.ctx.Done():
		return true

	default:
		return false
	}

}