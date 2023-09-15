curl --location 'localhost:9000/api/v1/employee/create' \
--header 'Cookie: x-name=Harry Potter' \
--header 'Content-Type: application/json' \
--data '{"bio": "the protagonist of the book"}'
curl --location 'localhost:9000/api/v1/employee/create' \
--header 'Cookie: x-name=Hermione Granger' \
--header 'Content-Type: application/json' \
--data '{"bio": "the know it all"}'
curl --location 'localhost:9000/api/v1/employee/read?id=3'
curl --location 'localhost:9000/api/v1/employee/read?id=2'
curl --location --request PATCH 'localhost:9000/api/v1/employee/update?id=1&vacationDays=3' \
--header 'Content-Type: application/json' \
--data '{"bio": "member of the order of the phoenix"}'
curl --location 'localhost:9000/api/v1/employee/read?id=1'