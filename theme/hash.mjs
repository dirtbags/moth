// Dan Bernstein hash v1
// Used until MOTH v3.5
function djb2(buf) {
    let h = 5381
    for (let c of (new TextEncoder()).encode(buf)) { // Encode as UTF-8 and read in each byte
      // JavaScript converts everything to a signed 32-bit integer when you do bitwise operations.
      // So we have to do "unsigned right shift" by zero to get it back to unsigned.
      h = (((h * 33) + c) & 0xffffffff) >>> 0
    }
    return h
  }
  
  // Used until MOTH v4.5
  async function sha256(message) {
    const msgUint8 = new TextEncoder().encode(message);                           // encode as (utf-8) Uint8Array
    const hashBuffer = await crypto.subtle.digest('SHA-256', msgUint8);           // hash the message
    const hashArray = Array.from(new Uint8Array(hashBuffer));                     // convert buffer to byte array
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join(''); // convert bytes to hex string
    return hashHex;
  }