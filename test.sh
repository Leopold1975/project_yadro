# admin password: admin
# user1 password: 123456
# user2 password: 1234

user_token=$(curl -s -D - -X POST http://localhost:4444/login \
    -H "Content-Type: application/json" \
    -d '{"username": "user1", "password": "123456"}'  \
    | grep -F "Authorization" | sed 's/Authorization: Bearer //' | tr -d '\r')

admin_token=$(curl -s -D - -X POST http://localhost:4444/login \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin"}' \
    | grep -F "Authorization" | sed 's/Authorization: Bearer //' | tr -d '\r')

# POST /update с admin токеном
curl -X POST http://localhost:4444/update \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $admin_token" \
    -d "{}" && \

# POST /update с user токеном (некорректный запрос)
curl -X POST http://localhost:4444/update \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $user_token" \
    -d '{}' && \

# GET /pics?search="I'll follow your questions" с user токеном
curl -X GET "http://localhost:4444/pics?search=I'll%20follow%20your%20questions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $user_token" && \

# GET /pics?search="apple doctor" с admin токеном
curl -X GET "http://localhost:4444/pics?search=apple%20doctor" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $admin_token" || \

echo "Tests failed"