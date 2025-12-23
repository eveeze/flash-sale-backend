import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// KONFIGURASI SERANGAN
export const options = {
  // Skenario: Ramp up (pemanasan) -> Serangan Puncak -> Reda
  stages: [
    { duration: '5s', target: 100 }, // Naik ke 100 user dalam 5 detik
    { duration: '10s', target: 1000 }, // BOM! Naik ke 1.000 user
    { duration: '10s', target: 1000 }, // Tahan di 1.000 user
    { duration: '5s', target: 0 }, // Turun pelan-pelan
  ],
};

export default function () {
  const url = 'http://localhost:8080/purchase';

  // Data Payload
  // Kita random User ID biar seolah-olah beda orang
  const payload = JSON.stringify({
    user_id: randomIntBetween(1, 100000),
    product_id: 1, // Pastikan ini ID barang yang stoknya 5.000 tadi
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  // TEMBAK!
  const res = http.post(url, payload, params);

  // Validasi: Kita anggap sukses kalau statusnya 202 (Accepted)
  // Kalau stok habis (409 Conflict), itu wajar, tapi sistem gak boleh error 500
  check(res, {
    'is status 202': (r) => r.status === 202,
    'is status 409 (sold out)': (r) => r.status === 409,
    'is not error 500': (r) => r.status !== 500,
  });

  // Jeda dikit (0.1 detik) biar gak kena rate limit OS lokal (optional)
  sleep(0.1);
}
