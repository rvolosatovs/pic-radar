# General
I built an web app, which searches imgur.com for random top 100 images corresponding to the query you type and shows a random one. 
It is hosted at [http://2id60.win.tue.nl/~s141426](http://2id60.win.tue.nl/~s141426/).
I have tested the website on `Chromium 62.0.3202.94` and `Mozilla Firefox 57.0.2`.

# Design
The app is built mostly using React.I used Bootstrap library for React to design the website and also FontAwesome for icons at the footer. Bootstrap helped make the website responsive and better-looking.
I some custom CSS for the app, i.e. I made `<main>` element respect size of header and footer, added colors to divide sections and impove the looks of the website. I have added an outline button, which changes color, once it's hovered upon.
See screenshots at the end of the document.

# Interaction
Once the user types in the search bar, requests is sent to the API to query for the user specified input. There is a small(50ms) delay between last search bar content change and executing the request to avoid sending requests while user is still typing the queries. The app initally searches for 'hello AND world', to avoid blank image on page load. 
If no results are found, a "image-not-found" placeholder is shown.
To workaround the issue with CORS not being supported by Imgur API I have developed a small backend in Go, which is also running on university server. It is available on port `8042` and provides one REST endpoint `/data`, which takes a `q` parameter and passes that to Imgur API. This is then called from front-end using the `Fetch` API.
`jQuery` was not required given the capabilities of `React` and `Fetch`.

# Backend
I used Go for the backend. No framework is required, given a very good standard library. I used InfluxDB database for analytics(storing the queries) and embedded BoltDB for storing user data.

# Databases
I used InfluxDB database for analytics(storing the queries) and embedded BoltDB for storing user data. The user can register/login and specify his favorite query, for this data I use BoltDB(login,password,query). InfluxDB stores the interaction log(time,duration of query,endpint,query,user,address)

# API
- `/image` returns a JSON array of JSON-encoded images as defined by Imgur API. Queries to this endpoint must be of "GET" method and should include a `q` parameter, which specifies the search query. Once a request to this endpoint is received, the specified search query is sent to Imgur and the result(if successful) is sent to the client.
More info can be found at [https://apidocs.imgur.com/#](https://apidocs.imgur.com/#)

- `/register` returns a JSON-encoded user object. Queries to this endpoint must be of "POST" method and should include JSON-encoded user object in the request body. This endpoint is used to register a user. Length of password and login should be at least 2.
input : 
```
{
    'login': ...,
    'password': ...,
    'query': ...
}
```

output : 
```
{
    'login': ...,
    'password': ...,
    'query': ...
}
```

- `/login` returns a JSON-encoded user object. Queries to this endpoint must be of "POST" method and should include JSON-encoded user object in the request body.

input : 
```
{
    'login': ...,
    'password': ...,
    'query': ...
}
```

output : 
```
{
    'login': ...,
    'password': ...,
    'query': ...
}
```

# Screenshots
![](/home/rvolosatovs/pictures/scrot/2017-12-22-11:37:31-screenshot.png)
![](/home/rvolosatovs/pictures/scrot/2017-12-22-11:35:26-screenshot.png)
![](/home/rvolosatovs/pictures/scrot/2017-12-22-11:51:40-screenshot.png)
