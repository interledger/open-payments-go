# signed request examples

`contextToRequestLike(ctx)`

Good (from bruno client)

```json
happy-life-auth-1      | {
happy-life-auth-1      |   clientKey: {
happy-life-auth-1      |     kid: 'keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5',
happy-life-auth-1      |     x: 'ubqoInifJ5sssIPPnQR1gVPfmoZnJtPhTkyMXNoJF_8',
happy-life-auth-1      |     alg: 'EdDSA',
happy-life-auth-1      |     kty: 'OKP',
happy-life-auth-1      |     crv: 'Ed25519'
happy-life-auth-1      |   },
happy-life-auth-1      | {
happy-life-auth-1      |   'contextToRequestLike(ctx)': {
happy-life-auth-1      |     url: 'http://localhost:4006/',
happy-life-auth-1      |     method: 'POST',
happy-life-auth-1      |     headers: {
happy-life-auth-1      |       accept: 'application/json, text/plain, */*',
happy-life-auth-1      |       'content-type': 'application/json',
happy-life-auth-1      |       'content-digest': 'sha-512=:/LfvBez/1knzYV3v4+Ej1qidX28IuoPp4jJBNSTkgBAu5TN5qS2FrfEWJohbBjIk1Xg7+qanR6VPm2+XyrZ3lQ==:',
happy-life-auth-1      |       signature: 'sig1=:0OiXIX1o5T7qb8uEzPzmde1mTTfJvLsheof5x7n+gSu8DILeQmwWQJ0uCEEZOL3mYH9ivv0zOWICdxucm/SxDQ==:',
happy-life-auth-1      |       'signature-input': 'sig1=("@method" "@target-uri" "content-digest" "content-length" "content-type");created=1747680150;keyid="keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5";alg="ed25519"',
happy-life-auth-1      |       'request-start-time': '1747680150717',
happy-life-auth-1      |       'user-agent': 'axios/1.6.7',
happy-life-auth-1      |       'content-length': '160',
happy-life-auth-1      |       'accept-encoding': 'gzip, compress, deflate, br',
happy-life-auth-1      |       host: 'localhost:4006',
happy-life-auth-1      |       connection: 'close'
happy-life-auth-1      |     },
happy-life-auth-1      |     body: '{"access_token":{"access":[{"type":"incoming-payment","actions":["create","read","list","complete"]}]},"client":"https://happy-life-bank-backend/accounts/pfry"}'
happy-life-auth-1      |   }
happy-life-auth-1      | }
```

Bad (from go client, httpsign impl)

```json
happy-life-auth-1      | {
happy-life-auth-1      |   'contextToRequestLike(ctx)': {
happy-life-auth-1      |     url: 'http://localhost:4006/',
happy-life-auth-1      |     method: 'POST',
happy-life-auth-1      |     headers: {
happy-life-auth-1      |       host: 'localhost:4006',
happy-life-auth-1      |       'user-agent': 'Go-http-client/1.1',
happy-life-auth-1      |       'content-length': '160',
happy-life-auth-1      |       'content-digest': 'sha-512=:SW/Vv1Dicc/1lqislo8sip5sMvtjAsgzgv3ZHnukSXGE5lxwbL5HWKlDnNB7lLWUKD+rC52JdxObeNJrwEa+Rw==:',
happy-life-auth-1      |       'content-type': 'application/json',
happy-life-auth-1      |       signature: 'sig1=:nThEre0AxJmc+m67mMZlkmERiYc3cdDtAL4jfhBgN4aOg6bmlmqhUjA+5xZT/Cq6vD4w0Kp35TC3Fsi47PNpBg==:',
happy-life-auth-1      |       'signature-input': 'sig1=("@method" "@target-uri" "content-digest" "content-length" "content-type");created=1747680669;alg="ed25519";keyid="keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5"',
happy-life-auth-1      |       'accept-encoding': 'gzip'
happy-life-auth-1      |     },
happy-life-auth-1      |     body: '{"access_token":{"access":[{"actions":["create","read","list","complete"],"type":"incoming-payment"}]},"client":"https://happy-life-bank-backend/accounts/pfry"}'
happy-life-auth-1      |   }
happy-life-auth-1      | }
```

Bad (from go client, manual impl)

```json
happy-life-auth-1      |   'contextToRequestLike(ctx)': {
happy-life-auth-1      |     url: 'http://localhost:4006/',
happy-life-auth-1      |     method: 'POST',
happy-life-auth-1      |     headers: {
happy-life-auth-1      |       host: 'localhost:4006',
happy-life-auth-1      |       'user-agent': 'Go-http-client/1.1',
happy-life-auth-1      |       'content-length': '160',
happy-life-auth-1      |       accept: 'application/json, text/plain, */*',
happy-life-auth-1      |       'accept-encoding': 'gzip, compress, deflate, br',
happy-life-auth-1      |       'content-digest': 'sha-512="SW/Vv1Dicc/1lqislo8sip5sMvtjAsgzgv3ZHnukSXGE5lxwbL5HWKlDnNB7lLWUKD+rC52JdxObeNJrwEa+Rw=="',
happy-life-auth-1      |       'content-type': 'application/json',
happy-life-auth-1      |       signature: 'sig1=:Ln/J3dJxOxe/PyLcrzzIhZ1hp3lrxJ9qY8lUd1TJAqROfM2uZ7UhNBjyiGlCok1lsjTzn4HFxBBnmirlX4WBDw==:',
happy-life-auth-1      |       'signature-input': 'sig1=("@method" "@target-uri" "content-digest" "content-length" "content-type");created=1747794637;keyid="keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5";alg="ed25519"'
happy-life-auth-1      |     },
happy-life-auth-1      |     body: '{"access_token":{"access":[{"actions":["create","read","list","complete"],"type":"incoming-payment"}]},"client":"https://happy-life-bank-backend/accounts/pfry"}'
happy-life-auth-1      |   }
happy-life-auth-1      | }
```

Bad, from go client httpsfv impl

```json
happy-life-auth-1      |   clientKey: {
happy-life-auth-1      |     kid: 'keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5',
happy-life-auth-1      |     x: 'ubqoInifJ5sssIPPnQR1gVPfmoZnJtPhTkyMXNoJF_8',
happy-life-auth-1      |     alg: 'EdDSA',
happy-life-auth-1      |     kty: 'OKP',
happy-life-auth-1      |     crv: 'Ed25519'
happy-life-auth-1      |   },
happy-life-auth-1      |   'contextToRequestLike(ctx)': {
happy-life-auth-1      |     url: 'http://localhost:4006/',
happy-life-auth-1      |     method: 'POST',
happy-life-auth-1      |     headers: {
happy-life-auth-1      |       host: 'localhost:4006',
happy-life-auth-1      |       'user-agent': 'Go-http-client/1.1',
happy-life-auth-1      |       'content-length': '160',
happy-life-auth-1      |       accept: 'application/json, text/plain, */*',
happy-life-auth-1      |       'accept-encoding': 'gzip, compress, deflate, br',
happy-life-auth-1      |       'content-digest': 'sha-512=:SW/Vv1Dicc/1lqislo8sip5sMvtjAsgzgv3ZHnukSXGE5lxwbL5HWKlDnNB7lLWUKD+rC52JdxObeNJrwEa+Rw==:',
happy-life-auth-1      |       'content-type': 'application/json',
happy-life-auth-1      |       signature: 'sig1=:RIJ9zTje3zN0060vzCLNOmHI/LmfHsG1GB95vtGks1EykkNEIx83MwHrOi8AFERdh0KAirI5eKvLnnJ45EJaAQ==:',
happy-life-auth-1      |       'signature-input': 'sig1=("@method" "@target-uri" "content-digest" "content-length" "content-type");created=1747794758;keyid="keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5";alg="ed25519"'
happy-life-auth-1      |     },
happy-life-auth-1      |     body: '{"access_token":{"access":[{"actions":["create","read","list","complete"],"type":"incoming-payment"}]},"client":"https://happy-life-bank-backend/accounts/pfry"}'
happy-life-auth-1      |   }
happy-life-auth-1      | }
```
