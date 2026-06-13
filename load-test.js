import http from 'k6/http';
import { check, sleep } from 'k6';

// Configure the load test: 50 concurrent users for 30 seconds
export const options = {
  vus: 50, 
  duration: '30s',
};

export default function () {
  // Hit the GET endpoint. 
  // Because the rate limit is 10/sec, almost all of these will be 429s.
  // That's exactly what we want to test!
  let res = http.get('http://localhost:8080/users/1');
  
  // Verify the server responds with either 200 (allowed) or 429 (rate limited)
  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
  });
}
