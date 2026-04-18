import CryptoJS from 'crypto-js'

const ENCRYPTION_KEY = 'So722DGlBzRASCG9so/Knfuy81cdhQ7gArgZjcJHvpQ='

export function encryptPassword(password) {
  if (!password) return ''
  const key = CryptoJS.enc.Base64.parse(ENCRYPTION_KEY)
  const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
  
  const encrypted = CryptoJS.AES.encrypt(password, key, {
    iv: iv,
    mode: CryptoJS.mode.CBC,
    padding: CryptoJS.pad.Pkcs7
  })
  return encrypted.ciphertext.toString(CryptoJS.enc.Base64)
}

export function decryptPassword(encryptedPassword) {
  if (!encryptedPassword) return ''
  try {
    const key = CryptoJS.enc.Base64.parse(ENCRYPTION_KEY)
    const iv = CryptoJS.enc.Utf8.parse('0000000000000000')
    
    const cipherParams = CryptoJS.lib.CipherParams.create({
      ciphertext: CryptoJS.enc.Base64.parse(encryptedPassword)
    })
    const decrypted = CryptoJS.AES.decrypt(cipherParams, key, {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7
    })
    return decrypted.toString(CryptoJS.enc.Utf8)
  } catch (e) {
    console.error('Decryption failed:', e)
    return ''
  }
}