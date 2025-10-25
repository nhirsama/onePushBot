package auth

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/clearsign"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

type VerifyResult struct {
	Valid    bool
	SignedAt int64  //签名时间戳
	Message  []byte //签名信息
	Err      error
	Hash     string //16位哈希摘要
}

// VerifyPGPSignature 从原始文本中提取并验证PGP签名
func (a *Auth) VerifyPGPSignature(signedMessage []byte) VerifyResult {
	result := VerifyResult{
		Valid: false,
		Err:   nil,
	}

	//解析 Clear-Signed 消息
	block, _ := clearsign.Decode(signedMessage)
	if block == nil {
		result.Err = errors.New("failed to decode PGP clear-signed message (invalid format or missing block)")
		return result
	}

	// openpgp.CheckDetachedSignature 验证消息内容 block.Bytes 是否由签名 block.ArmoredSignature.Body 签署
	sigBytes, err := io.ReadAll(block.ArmoredSignature.Body)
	if err != nil {
		result.Err = fmt.Errorf("failed to read signature body: %w", err)
		return result
	}
	md, err := openpgp.CheckDetachedSignature(
		a.keyRing,
		bytes.NewReader(block.Bytes), // 要验证的消息内容
		bytes.NewReader(sigBytes),    // 签名数据
		nil,
	)

	if err != nil {
		result.Err = fmt.Errorf("signature verification failed: %w", err)
		return result
	}

	// 3. 验证成功，提取信息
	if md != nil {
		result.Valid = true
		// 原始消息体在 block.Bytes 中
		result.Message = block.Bytes

		// 在 CheckDetachedSignature 的成功返回值 md (*openpgp.Entity) 中，
		// 签名时间位于 PGP 签名数据包（Signature Packet）内部。
		// 在 clear-signed 消息中，签名数据位于 block.ArmoredSignature.Body 中。

		// 需要从 block 中解析出签名数据包，然后从中提取创建时间。
		signaturePacket, err := packet.Read(bytes.NewReader(sigBytes))
		if err != nil {
			result.Valid = false
			result.Err = fmt.Errorf("failed to read signature packet: %w", err)
			return result
		}

		sig, ok := signaturePacket.(*packet.Signature)
		if !ok {
			result.Valid = false
			result.Err = errors.New("expected a signature packet, got a different packet type")
			return result
		}
		hashTag16bit := sig.HashTag

		// 将其格式化为 4 个十六进制字符的字符串，例如 "361F"
		result.Hash = fmt.Sprintf("%04X", hashTag16bit)
		// 提取签名时间
		result.SignedAt = sig.CreationTime.Unix()
		return result
	} else {
		// md 应该不会是 nil，因为 err 检查已经通过。如果 md 为 nil 且 err 也为 nil，则可能是逻辑错误。
		result.Err = errors.New("verification succeeded, but signer entity is unexpectedly nil")
		return result
	}
}
