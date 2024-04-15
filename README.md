localhost:8081/api/v1/user_banner?tag_id=1&feature_id=1 GET
localhost:8081/api/v1/user_banner?tag_id=3&feature_id=3 GET
localhost:8081/api/v1/user_banner?tag_id=1&feature_id=1&use_last_revision=true GET

localhost:8081/api/v1/banner?feature_id=2&tag_id=2&limit=10&offset=0 GET

localhost:8081/api/v1/banner?feature_id=1&tag_id=2&limit=10&offset=0 GET



localhost:8081/api/v1/banner POST
Body:
{
   "tag_ids": [
      2
   ],
   "feature_id": 4,
   "content": "{\"content\": \"Banner 3 - Version 1\"}"
}

localhost:8081/api/v1/banner/12 PATCH
Body:
{
   "tag_ids": [
      1
   ],
   "feature_id": 4,
   "content": "{\"content\": \"Banner 4  - Version 1\"}"
}


localhost:8081/api/v1/banner/12 DELETE
Body:
{
   "tag_ids": [
      1
   ],
   "feature_id": 4,
   "content": "{\"content\": \"Banner 4  - Version 1\"}"
}


localhost:8080/signup

localhost:8080/signin

localhost:8080/logout

{
   "login": "test",
   "password": "test"
}
