 kubectl -n openfaas-fn create secret generic secrets --from-file=./key --from-file=./key.pub --from-file=./serviceAccountKey.json \
  --from-literal="ts-client-secret=$CLIENT_SECRET" --from-literal="mongo-host=$MONGO_HOST" \
  --from-literal="mongo-database=$MONGO_DB" \
  --from-literal="recaptcha-key=$RECAPTCHA_KEY" --from-literal="ref-email-pass=$REF_EMAIL_PASS" \
  --from-literal="redis-pwd=$REDIS_PWD" --from-literal="admin-username=$ADMIN_USERNAME"\
  --from-literal="admin-password=$ADMIN_PASSWORD" --from-literal payload-secret="$PAYLOAD_SECRET"\
  --from-literal="phone-auth-token=$PHONE_AUTH_TOKEN" --from-literal="phone-auth-id=$PHONE_AUTH_ID" 