local page_403 = [[<!DOCTYPE html>
<html>
<head>
<meta content="text/html;charset=utf-8" http-equiv="Content-Type">
<meta content="utf-8" http-equiv="encoding">
<title>403 Forbidden</title>
<style>
    body {
        width: 40em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>403 Forbidden</h1>
<p>当前访问可能对网站造成威胁，已被阻断，如有疑问请联系网站管理员并提供请求ID。</p>
<p><em>您的请求ID为：{{REQUEST_ID}}</em></p>
</body>
</html>]]


return page_403