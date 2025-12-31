package vna

import "fmt"

func (v *VNA) Boot() error {

	// ============ SET UDP CONNECTION ============
	if err := v.InitConnection(); err != nil {
		
		return fmt.Errorf("UDP init connection failed %v", err)
	}

	// ============ CREATE CryptoKeys Struct and load it in vna ============
	if err := v.LoadCryptoKeys(); err != nil {
        return fmt.Errorf("crypto init: %w", err)
    }

	if err := v.RequestIP(); err != nil {
        return fmt.Errorf("ip request: %w", err)
    }


	if err := v.SetupAdapter();err !=nil{

		return fmt.Errorf("adatper setup error %w",err)
	}

	return nil


}