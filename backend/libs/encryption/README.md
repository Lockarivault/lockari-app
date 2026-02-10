# Encryption Library Usage Guide

Esta biblioteca fornece ferramentas para gestão de metadados e operações criptográficas utilizando o padrão de **Envelope Encryption** (KEK e DEK).

## 1. Gestão de Metadados (`EncryptMetadata`)

A struct `EncryptMetadata` armazena o estado e a configuração de uma chave (KEK ou DEK). Ela não contém o material da chave em si, apenas informações sobre ela.

### Criar Metadado
```go
import "github.com/Lockarivault/lockari-app/backend/libs/encryption"

// Criando um novo metadado para uma DEK
metadata := encryption.NewEncryptMetadata(
    "key-uuid-123",           // ID único
    encryption.KeyTypeDEK,    // Tipo (KEK, DEK, ROOT)
    "AES_GCM_256",            // Algoritmo
    "AWS_KMS",                // Provedor
)

// Configurando campos adicionais (Fluent API)
metadata.WithParentKeyID("kek-id-456").
    WithFingerprint("hash-da-chave").
    WithCreatedBy("user-id-789")

// Validação
if err := metadata.Validate(); err != nil {
    log.Fatal(err)
}
```

### Atualizar Estado
```go
metadata.MarkAsUsed()     // Atualiza data de último uso
metadata.MarkAsRotated()  // Atualiza data de rotação e timestamp de update
metadata.WithStatus(encryption.StatusDisabled) // Desativa a chave
```

---

## 2. Operações com KEK e DEK (`Encryptor`)

O `Encryptor` lida com a geração de chaves de dados e sua proteção usando uma chave mestra (KEK).

### Inicializar o Encryptor
```go
// Você precisa do material da KEK (32 bytes para AES-256)
kekMaterial := []byte("sua-chave-mestra-de-32-bytes-!!") 
keyID := "id-da-kek"

enc, err := encryption.NewEncryptor(kekMaterial, keyID)
if err != nil {
    log.Fatal(err)
}
```

### Fluxo de Trabalho (Ciclo de Vida)

#### A. Gerar uma nova DEK
Gera uma chave aleatória para criptografar seus dados sensíveis.
```go
dek, err := enc.GenerateDEK(ctx)
```

#### B. Proteger a DEK (Criptografar com KEK)
Gera um `Envelope` contendo o ciphertext da DEK e o Nonce. É isso que você salva no banco de dados.
```go
envelope, err := enc.EncryptDEKWithKEK(ctx, dek)
// envelope.Ciphertext -> DEK criptografada (Base64)
// envelope.Nonce      -> IV usado (Base64)
```

#### C. Recuperar a DEK (Descriptografar)
Usa a KEK para descriptografar o envelope e obter a DEK original.
```go
originalDEK, err := enc.DecryptDEK(ctx, envelope)
```

---

## 3. Utilitários de Encoding

Helpers simples para lidar com strings Base64 URL-Safe e Hexadecimal.

```go
str := encryption.EncodeBase64(data)
data, err := encryption.DecodeBase64(str)

hexStr := encryption.EncodeHex(data)
data, err := encryption.DecodeHex(hexStr)
```
