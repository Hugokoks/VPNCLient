package vna

func (v *VNA) CtxStopped() bool {
	select {
	case <-v.ctx.Done():
		return true

	default:
		return false
	}

}