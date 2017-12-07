# General
I built an web app, which searches imgur.com for random top 100 images corresponding to the query you type and shows a random one. 
It is hosted at [http://2id60.win.tue.nl/~s141426](http://2id60.win.tue.nl/~s141426/).
I have tested the website on `Chromium 62.0.3202.94` and `Mozilla Firefox 56.0.2`.

# Design
The app is built mostly using React.I used Bootstrap library for React to design the website and also FontAwesome for icons at the footer. Bootstrap helped make the website responsive and better-looking.
I some custom CSS for the app, i.e. I made `<main>` element respect size of header and footer, added colors to divide sections and impove the looks of the website. I have added an outline button, which changes color, once it's hovered upon.
See screenshots at the end of the document.

# Interaction
Once the user types in the search bar, requests is sent to the API to query for the user specified input. There is a small(50ms) delay between last search bar content change and executing the request to avoid sending requests while user is still typing the queries. The app initally searches for 'hello AND world', to avoid blank image on page load. 
If no results are found, a "image-not-found" placeholder is shown.
To workaround the issue with CORS not being supported by Imgur API I have developed a small backend in Go, which is also running on university server. It is available on port `8042` and provides one REST endpoint `/data`, which takes a `q` parameter and passes that to Imgur API. This is then called from front-end using the `Fetch` API.
`jQuery` was not required given the capabilities of `React` and `Fetch`.

# Screenshots
![](/home/rvolosatovs/pictures/scrot/2017-12-07-16:08:55-screenshot.png)
![](/home/rvolosatovs/pictures/scrot/2017-12-07-16:09:53-screenshot.png)
