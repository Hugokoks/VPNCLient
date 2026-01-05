package vna

import (
	"VPNClient/keys"
	"fmt"
)

func (v *VNA) Boot() error {
	

	var cryptoKeys keys.CryptoKeys

	if err := cryptoKeys.LoadOrCreateClientIdentity(); err != nil{

		return fmt.Errorf("failed to obtain clientPub and clientPriv %w",err)

	}

	if err := cryptoKeys.LoadServerPublicKey(); err != nil{

		return fmt.Errorf("failed to obtain serverPub %w",err)
	}
	
	v.Keys = cryptoKeys

	// ============ SET UDP CONNECTION ============
	if err := v.InitConnection(); err != nil {
		
		return fmt.Errorf("UDP init connection failed %v", err)
	}
	

	if err := v.RequestIP(); err != nil {
        return fmt.Errorf("ip request: %w", err)
    }


	if err := v.SetupAdapter();err !=nil{

		return fmt.Errorf("adatper setup error %w",err)
	}

	return nil


}