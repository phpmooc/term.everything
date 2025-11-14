# How I Did It!

A long long time ago, when a program wanted to draw something to the screen, it
could just write to a certain spot in RAM. The screen was a wild west and anyone
could draw anywhere anytime.

Nowadays, display access is coordinated by a special program called the
[Display Server](https://en.wikipedia.org/wiki/Windowing_system#Display_server),
which "coordinates the input and output of its clients to and from the rest of
the operating system, the hardware and each other". With input being mouse/keyboard/etc
and output being graphics, images, anything you see in a "window".

On modern linux, for display servers most systems use either the [Wayland (protocol)](https://en.wikipedia.org/wiki/Wayland_(protocol)) or [x11](https://en.wikipedia.org/wiki/X_Window_System).
We're going to focus on Wayland, mostly because it's newer and there are systems in place
to run older x11 apps in wayland for backwards compatibility.

## Wayland (protocol)
Notice that it's Wayland *protocol* not wayland display-server. That's because
Wayland is not a display server. It's a [protocol](https://en.wikipedia.org/wiki/Communication_protocol); it'a a way for programs to communicate
to a hypothetical display server. When people say that their system is running 
Wayland, what they really mean is that their system is running a display server
that speaks the wayland protocol.

What's even better
is that Wayland does not have a set [Rendering Model](https://en.wikipedia.org/wiki/Wayland_(protocol)#Rendering_model), the programs decide entirely on their own how they
will draw their "windows" or other graphics, then they just hand it over to the
display server.

## How to draw
The wayland protocol actually has no opinion on what you do with the graphics the
programs give you. Sure, they probably expect that you will use [Kernel Mode Setting](https://en.wikipedia.org/wiki/Mode_setting), but really if you wanted to, you could print
out the graphics and give them to a league of crochet grandmas to individually
tie together every single pixel into the afghan of legend! The programs would never be the wiser.

Again, the main benefit of wayland is that the programs don't care
how they get input as long as it *gets input*, and it doesn't care what happens to
its output.

So, this means we can do anything we want, so let's output to the terminal!

I take the output given to us by the client and convert the images to terminal
output, the utf8 characters with ansi escape codes via the [chafa library](https://github.com/hpjansson/chafa/). For input, I take the keyboard and mouse (yes terminals support mice)
from the stdin. And that's it! Of course, there are about 10K lines of code needed to actually do this in practice, but if you're interested in that I invite you to [check out the source code](../termeverything/).

## What else can you do with wayland
I have many other crazy ideas of what else to do custom wayland display server, so stay tuned.
