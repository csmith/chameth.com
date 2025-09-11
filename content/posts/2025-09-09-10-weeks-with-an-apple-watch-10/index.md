---
date: 2025-09-09
title: "10 Weeks with an Apple Watch 10"
permalink: /10-weeks-with-an-apple-watch-10/
tags: [hardware, personal, health]
resources:
  - src: watch.jpg
    name: "An Apple Watch 10 being worn, with a blue analogue clock on the display, and icons/data shown in the corners"
    title: "My watch. Yes, I am available for wrist modelling opportunities."
  - src: sleep.png
    name: "A graph showing sleep phases over time. There's a noticeable transition form patchy data to more smooth data."
    title: "Sleep data; spot when I swapped devices!"
opengraph:
  title: "10 Weeks with an Apple Watch 10 · Chameth.com"
  type: article
  url: /10-weeks-with-an-apple-watch-10/
  image: /10-weeks-with-an-apple-watch-10/watch.jpg
---

{% figure "right" "An Apple Watch 10 being worn, with a blue analogue clock on the display, and icons/data shown in the corners" %}

Around ten weeks ago[^1] I picked up an Apple Watch 10, and have been wearing
it almost constantly since. It's not my first Apple Watch — I had a Series 5
for a bit back in 2020 — but it's the first time I've actually stuck with it.
Ten weeks seems like an apt time to reflect on it.

Firstly, why did I even bother? Well, for a couple of years I'd been wearing
a Xiaomi Smart Band 7, mainly to monitor my sleep stats and set alarms that
won't wake up everyone else nearby. Its battery life was fantastic — with
notifications and other things turned off, I got about a month of use between
charges — but actually using it felt like trying to order food via the medium
of interpretive dance.

My biggest gripe was the screen lock. If I didn't have the screen locked then
I'd periodically trigger it during the night when I moved around
and it came in contact with my chest or leg. With the lock enabled you had to
deliberately swipe up from bottom to top to enable interaction, but it
didn't work reliably. When I wanted to adjust an alarm, I'd be stood swiping
repeatedly trying to get it to respond. When you finally get it unlocked, the
whole interface is just _fiddly_.

The other issue was the data quality. There were some nights when I'd been
woken up, sometimes even getting up and moving around, and it just didn't show
it in the data. If it can't even get whether I'm asleep right, can I trust
anything else it says?

I spent a while researching the best devices for sleep tracking. The Oura ring
came highly recommended, but it was expensive and required a subscription to
do anything useful. No thanks! The Apple Watch was consistently rated pretty
well, and I reasoned I could pick up a refurbished older unit. I've been
[using an iPhone as my daily driver](/apple-google-aligned-incentives/) for
a while, so it'd fit right into my begrudged walled garden.

The Series 10 has a significant advantage, though: it charges much quicker than
all the previous generations. On a 30-minute charge, the Series 10 can go from
0 to 60%; the 9 can only make it to 40%, and my old 5 a measly 30%. Shorter
charge times means I'm far less likely to leave it on charge and wander off
without it. In some ways the daily charging is more convenient than monthly: the
wireless charger sits on my desk, and I plop the watch on it for a little while
in the evening; I don't need to dig out the weird pogo-pin connector that has
vanished sometime in the last four weeks, then carefully arrange it so it stays
attached.

### How a watch maybe saved my life

One of the big features of the Apple Watch, like many other wearable devices,
is health and fitness tracking. I didn't think much about this, beyond the
sleep data I wanted, at first. I've never had a particularly good relationship
with exercise[^2], but I do like some good statistics. I started going for
walks more often to get more data and see the graphs of VO2 max and HR recovery
gradually inch up. That wasn't the most profound effect on my health, though…

<!--more-->

The recent versions of the Apple Watch have a feature that monitors for sleep
apnea, a disorder where you don't breathe properly during your sleep. I knew
I wasn't sleeping great — that's why I was paying attention to sleep tracking
data — but was still a bit surprised to get a notification from Apple Health
after wearing the watch for 30 days. It gives you a graph to print and take to
your doctor. So I diligently booked an appointment and a few weeks later went
to see my GP.

The appointment went about as you'd expect: talking about referral to a sleep
centre for a study, and so on. Towards the end, the doctor took my blood
pressure (sleep apnea can be caused by, and can cause high blood pressure, in
a lovely little vicious cycle). I don't think either of us expected anything;
it was just one of those standard checks for a related problem. After taking
the reading, the doctor looked at me and said, "I can't let you leave with a
BP like this". Yikes!

Blood pressure readings are split into stages: normal is under 120 mmHg[^3]
over 80 mmHg[^4], stage 1 is up to 140/90, stage 2 is up to 180/120,
and above is simply called "crisis". Guess where I was? Also, fun fact:
depending on how exactly you count and attribute things, hypertension comes out
as the single largest cause of death in humans. It doesn't kill you outright,
but contributes to strokes, coronary artery disease, heart failure, and lots
of other lovely things you don't want on your CV.

After doing a few more readings, it settled down to just under the "crisis"
stage and into the "holy crap, start treatment immediately" stage instead.
I won't labour on much more about this, but things are definitely moving in
the right direction now[^5].

While the Apple Watch didn't literally save my life, it triggered the chain of
events that led to this diagnosis and treatment. Who knows what would have
happened had it remained undetected? Probably nothing good. Also, go check your
blood pressure! It's simple to do and simple to treat if there is an issue,
but so many people are walking around with hypertension and not even realising.

{% update "2025-09-09" %}
Just a couple of hours after I published this post, Apple announced that they're
adding hypertension notifications in the new Series 11 watch. It looks like
it will also be supported in Series 9 and Series 10 watches. They expect it to
notify more than one million people of unknown hypertension in the first year.
{% endupdate %}

### Building gates in the walled garden

Even though the watch is arguably a life-saver, not everything is rosy.
It's an Apple product, so you're firmly locked in a walled garden, jealously
guarded by people wearing black turtlenecks. Fortunately, there are a few ways
to make it less painful.

All the health and fitness data is stored in Apple Health. You can export data
as a big messy file, but it's a very manual process to do so and the data format
is gnarly. Luckily, there's an app for that!
[Health Auto Export](https://www.healthyapps.dev/) can, well, automatically
export health data. It does what it says on the tin. It can send the data to
Home Assistant, over MQTT, or dump it in some cloud file services, but I just
have it post it all to a REST endpoint on a service I wrote. Then I dump it all
in a database and can do whatever I want with it![^6]

Another tool that's more useful on the watch itself is Apple Shortcuts. This
is their no-code "if-this-then-that"-type thing. You can make automations or
shortcuts that run a number of tasks. I have a whole slew of them I access
via a complication[^7] on my watch face: one which prompts for input and adds an
item to my to-do list (swiping to write letters is surprisingly not horrible),
one which lets me select from a bunch of pre-written ones ("charge kindle",
"take laundry out in 1 hour", type things), one to log my weight into Apple
Health, one which can open and close the blinds in my room, and so on. It's
a surprisingly robust and easy-to-use system and offers just enough freedom
that I'm not constantly grating on the edge of the walled garden.

Shortcuts being able to initiate arbitrary web requests is the real killer
feature for me. Anything I can't do on the watch itself, I can just farm off
to a web server and connect it up with a shortcut. No need to learn Swift or
pay Apple for the privilege of being a developer! For a lot of things, like
controlling the blinds or adding to-do items, I already had a HTTP endpoint
available and exposed over Tailscale[^8]. Adding it to the watch was just a case
of entering the right things in the Shortcuts app.

### Daily nitty-gritty

There are lots of other little bits and pieces that come up when using the watch
daily. I don't think I can bundle them up into a pleasing narrative arc, so
instead please enjoy some disjointed paragraphs of observations.

The Apple Watch has a lot of nagging abilities. It can notify you about your
fitness "rings", prompt you to stand up every hour, count how many seconds
you wash your hands for, etc. I think I'd dislike these just on general
principle, but the way it does them is _so_ condescending it's painful. I think
there's probably a cultural divide issue here, but there is no way in British
English to say "Great job! You washed your hands for 30 seconds!" without
it sounding like you're being amazingly sarcastic or like you're talking to a
young child. So I turned all of that nonsense off. It's meant to be a tool
not a wannabe life coach.

You can access Maps directly on the watch and even do navigation. It works
really well. The navigation mode has some nice haptic feedback: it does a short
pulse as you're approaching a turn, and then a long pulse at the actual turn.
I like it a lot more than having to dig out my phone or have the directions
read out. You get one pulse, glance down and see where you need to go, then it
reminds you a little later when it's time to do it. It's a delightful user
experience.

Watchfaces aren't quite so delightful. There's a limited number of built-in
ones, and some are "exclusive" to the Ultra Series, and you can't use them on
a peasant watch like a Series 10. Annoyingly, there's not one that does
exactly what I want: a plain analogue clock with an inset date and four
complication slots around the outside. Instead, I have to use one of the slots
to show the date. There are third-party watchfaces, but they have issues.
Firstly, there's no actual API for making watchfaces[^9], so what they do is
bodge it horribly by using a photo background that has fake widgets on it.
On top of that they're almost universally subscription-based. Again, no
thanks[^10].

The issue I had with the Smart Band triggering when I was sleeping is solved
trivially on an Apple Watch, by virtue of it having a physical button in the
crown. When you put it in sleep mode, you have to double-press the crown to
unlock it before it'll do anything else. It hasn't misfired once while I've
been using it.

{% figure "right" "A graph showing sleep phases over time. There's a noticeable transition form patchy data to more smooth data." %}

Finally, a special mention for the gesture controls. If you raise the watch
it reliably wakes up (switching from a low refresh rate, dimmed screen to
an active, brighter one); you can then double-tap your index finger and thumb
together and it will scroll down or page through whatever you're looking at.
The killer feature for this is navigating recipes: you can advance to the next
step in a recipe while your hands are covered in flour. It's also handy for
reading notifications: when one pops up, you can double-tap to expand it, then,
when it gets to the bottom, it'll outline the default button (often "Dismiss")
and you can double-tap again to click it.

### The verdict

I normally don't like writing an actual labelled conclusion, but it feels
like one is needed here! Overall, I'm happy with the watch. The daily
charging doesn't bother me, the data gathered seems reliable, the health
monitoring has obviously paid dividends already, and the walled garden isn't
_too_ chafing. It's a straight upgrade over my old Smart Band, and I think
it was worth the cost.

The original reason for getting the watch was better sleep tracking, though,
so how well did it do? I'm much happier with the data: it seems to more
accurately represent when I was awake in the night, and overall the sleep
phases just seem to make more sense. You can see in the graph that the old
data switched frequently between phases, and they didn't quite line up for
some reason; towards the right when the Apple Watch is supplying the data
instead there's a much more consistent pattern of sleep phases that repeat
over the course of the night.

I'm not going to advocate that you go out and buy one, though. I know my
requirements and usage aren't typical, and I've also not got experience with
any recent Android Wear alternatives or the new version of the Pebble watch
that's coming soon. You should definitely get your blood pressure checked,
though!

[^1]: OK, it's more like 14 now, it's taken a while for this post to make its
way from my brain into text.
[^2]: Exercise for the sake of exercising just seems so overwhelmingly
tedious and boring to me. And other types of exercise generally require
social interaction, coordinating with people, and so on.
[^3]: Who decided to use "millimetres of mercury displaced" as a unit? You
can't just put random chemical symbols in units! That's not how this works!
[^4]: Blood pressure readings have two parts: systolic (the pressure when the
heart is beating) and diastolic (the pressure between those beats). They're
generally presented with the systolic reading on top and the diastolic reading
below, and read as "X over Y". Now you know what some of the random numbers
they shout in medical shows mean! Yay learning!
[^5]: No thanks to how much salt is in everything. I'm pretty sure I've had at
least twice the recommended daily amount of salt in a single serving before.
Don't even get me started on the things that are "low salt" but are still full
of sodium from other sources. I don't have a problem with ionic compounds, I
have a problem with sodium!
[^6]: This mainly looks like drawing graphs that are slightly different to
the graphs in the Apple Health graphs, for reasons I'm not sure I can explain.
Making graphs is fun, OK?
[^7]: Complications are basically just home screen widgets, but with a fancy
name because they're on a watch.
[^8]: Tailscale actually causes me some problems here: everything works fine
when the watch is connected to my phone, as the phone handles the Tailscale
part, but if I'm not carrying my phone the watch will try to connect over
WiFi directly and doesn't understand anything about Tailscale. It happens
infrequently enough that I'll just live with it; it's not much worse than
having no signal on a phone.
[^9]: Yay walled gardens…
[^10]: I don't object to subscribing to things in general, but it has to be
something that's worth the ongoing cost and offers something in return for
the subscription. A watchface doesn't need enough ongoing maintenance to
justify subscribing to it, it's just a cash grab.