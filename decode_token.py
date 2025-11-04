import os, json, base64
token = os.environ.get('TOKEN')
if not token:
    print("Token not found in environment variable.")
else:
    try:
        payload = token.split('.')[1]
        payload += '=' * (-len(payload) % 4)
        decoded_payload = base64.urlsafe_b64decode(payload)
        print(json.dumps(json.loads(decoded_payload), indent=2))
    except Exception as e:
        print(f"An error occurred: {e}")
        print("Token payload: ", payload)