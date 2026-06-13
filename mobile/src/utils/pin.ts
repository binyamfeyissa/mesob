import * as SecureStore from "expo-secure-store";

const PIN_KEY = "mesob_pin_hash";

// Very simple deterministic derivation — prevents raw PIN from being stored in plaintext.
// Each digit is XOR'd with its position + a constant, then the result is hex-encoded.
// The security guarantee comes from SecureStore's AES-256 + device Keychain/Keystore.
function derive(pin: string): string {
  let out = "";
  for (let i = 0; i < pin.length; i++) {
    const byte = ((pin.charCodeAt(i) - 48) ^ (i * 17 + 137)) & 0xff;
    out += byte.toString(16).padStart(2, "0");
  }
  return out;
}

export async function storePin(pin: string): Promise<void> {
  await SecureStore.setItemAsync(PIN_KEY, derive(pin));
}

export async function verifyPin(pin: string): Promise<boolean> {
  const stored = await SecureStore.getItemAsync(PIN_KEY);
  if (!stored) return false;
  return stored === derive(pin);
}

export async function hasPinStored(): Promise<boolean> {
  const stored = await SecureStore.getItemAsync(PIN_KEY);
  return !!stored;
}
