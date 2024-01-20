Traffic
=======

Ever been sat in a traffic jam on the motorway and 
wondered how long you're going to be sat there? No?
Me neither since Google Maps/Waze/any live update SatNav
became popular. However whilst sat in the back of
a vehicle without access to one of those services I
decided to see if there were any websites offering
a similar service. And there are! But they are, frankly,
crap. Filled with ads, all trying to load a map and
lots of live update JavaScript. Very unhelpful when you
are on a stretch of motorway with only Edge/2G coverage.

Angered by this experience I built `traffic`.

It turns out that in the UK (or England at least), the
Highways Agency publishes an RSS feed of all the latest
scheduled and unscheduled traffic disruption. Unfortunately
this feed is for every motorway all at once.

`traffic` reads this feed and then breaks it down by the
road name (or, I guess number), creating a static HTML
website.

Just run this on a cron job every hour or so and never
worry about having a slow web connection when looking
for traffic updates again!