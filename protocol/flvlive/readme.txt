
flv直推
1. flv的header和第一个previous tag len 不要发送
2. 第一个字节magic number表示flv 推流，(暂时未定义)
3. 第一个tag是一个amf0的tag，"url":"rtmp://xx",表示为一个推流地址
4. 后续发送正常的tag+previous tag len
