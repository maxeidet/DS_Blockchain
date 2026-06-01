package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	Address    string
}

func NewWallet() (*Wallet, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	pub := PublicKeyBytes(&priv.PublicKey)
	addr := AddressFromPublicKey(pub)

	return &Wallet{
		PrivateKey: priv,
		PublicKey:  pub,
		Address:    addr,
	}, nil
}

func PublicKeyBytes(pub *ecdsa.PublicKey) []byte {
	// P-256 => 32 bytes per coordinate
	xBytes := pub.X.FillBytes(make([]byte, 32))
	yBytes := pub.Y.FillBytes(make([]byte, 32))

	out := make([]byte, 0, 65)
	out = append(out, 0x04) // uncompressed format
	out = append(out, xBytes...)
	out = append(out, yBytes...)
	return out
}

func AddressFromPublicKey(pubKey []byte) string {
	hash := sha256.Sum256(pubKey)
	return hex.EncodeToString(hash[:20])
}

func SignBytes(priv *ecdsa.PrivateKey, data []byte) (string, error) {
	hash := sha256.Sum256(data)

	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		return "", err
	}

	// fixed-width encoding: 32 bytes r + 32 bytes s
	rBytes := r.FillBytes(make([]byte, 32))
	sBytes := s.FillBytes(make([]byte, 32))

	sig := append(rBytes, sBytes...)
	return hex.EncodeToString(sig), nil
}

func VerifySignature(pubKeyBytes []byte, data []byte, sigHex string) (bool, error) {
	pubKey, err := ParsePublicKey(pubKeyBytes)
	if err != nil {
		return false, err
	}

	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, err
	}

	if len(sigBytes) != 64 {
		return false, errors.New("invalid signature length")
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	hash := sha256.Sum256(data)
	ok := ecdsa.Verify(pubKey, hash[:], r, s)
	return ok, nil
}

func ParsePublicKey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	// uncompressed P-256 = 1 + 32 + 32 = 65 bytes
	if len(pubKeyBytes) != 65 || pubKeyBytes[0] != 0x04 {
		return nil, errors.New("invalid public key format")
	}

	x := new(big.Int).SetBytes(pubKeyBytes[1:33])
	y := new(big.Int).SetBytes(pubKeyBytes[33:65])

	pub := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
		return nil, errors.New("public key is not on the curve")
	}

	return pub, nil
}