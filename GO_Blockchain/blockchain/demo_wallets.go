package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
)

func NewDeterministicWallet(seed string) (*Wallet, error) {
	curve := elliptic.P256()

	hash := sha256.Sum256([]byte(seed))
	d := new(big.Int).SetBytes(hash[:])

	// säkra att d ligger inom kurvans ordning
	n := curve.Params().N
	d.Mod(d, new(big.Int).Sub(n, big.NewInt(1)))
	d.Add(d, big.NewInt(1))

	priv := &ecdsa.PrivateKey{}
	priv.PublicKey.Curve = curve
	priv.D = d
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(d.Bytes())

	pub := PublicKeyBytes(&priv.PublicKey)
	addr := AddressFromPublicKey(pub)

	return &Wallet{
		PrivateKey: priv,
		PublicKey:  pub,
		Address:    addr,
	}, nil
}

func FaucetWallet() (*Wallet, error) {
	return NewDeterministicWallet("GO_BLOCKCHAIN_FAUCET_WALLET")
}