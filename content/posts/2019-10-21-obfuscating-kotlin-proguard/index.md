---
date: 2019-10-21
title: Obfuscating Kotlin code with ProGuard
description: In which Kotlin tries to be helpful, and we smite it
tags: [android, development]
format: long
permalink: /obfuscating-kotlin-proguard/

resources:
  - src: obfuscated.png
    name: Obfuscated code viewed in Android Studio
  - src: kotlin-proguard.png
    name: Kotlin and Proguard logos
    title: Kotlin + Proguard = fun

opengraph:
  image: /obfuscating-kotlin-proguard/obfuscated.png
---

{% figure "left" "Kotlin and Proguard logos" %}

Obfuscating code is the process of modifying source code or build output in
order to make it harder for humans to understand. It's often employed as a
tactic to deter reverse engineering of commercial applications or libraries
when you have no choice but to ship binaries or byte code. For Android apps,
[ProGuard](https://www.guardsquare.com/en/products/proguard) is part of the
default toolchain and obfuscation is usually only a config switch away.

I was recently working on an Android library written in Kotlin that my client
wanted obfuscated to try and protect some of their trade secrets that were
included. Not a problem, I thought: it's just a few lines of ProGuard config
and we're away. Four hours and lots of hair pulling later I finally got it
working...

<!--more-->

### If at first you don't succeed...

At first it seemed like ProGuard was refusing to obfuscate any class with a
`keep` rule. With a simple test class:

```kotlin
class Test {

    private val secret = 123
    var name = "Chris"

    fun greet() {
        println("Hi $name! Enter a number: ")
        readLine()?.let { guess(it.toInt()) }
    }

    private fun guess(attempt: Int) = println(
        if (attempt == secret) {
            "Correct!"
        } else {
            "Nope!"
        }
    )

}
```

And a ProGuard rule of:

```text
-keep public class Test {
    public void greet();
}
```

I expected that the `Test` class and the `greet` method would remain, but both
fields and the `guess` method would be obfuscated. When I built the project
and opened the class from Android Studio's APK inspector I was disappointed:

```kotlin
public final class Test public constructor() {
    public final var name: kotlin.String /* compiled code */

    private final val secret: kotlin.Int /* compiled code */

    public final fun greet(): kotlin.Unit { /* compiled code */ }

    private final fun guess(attempt: kotlin.Int): kotlin.Unit { /* compiled code */ }
}
```

Having your secret sauce in a field labelled "secret" isn't exactly the level of
obfuscation I was hoping for. ProGuard has lots of knobs that you can twist to
affect what it keeps and what it renames, but all the incantations of `-keep`,
`-keepmembernames`, `-allowobfuscation`, and so on, that I could come up with
either resulted in the class completely vanishing (because it wasn't kept) or
showing up with all its symbols intact.

There are lots of useful Stack Overflow posts describing how to obfuscate
a single method, or keep a single method and obfuscate the rest, but nothing
I tried seemed to make a difference. I'm not the biggest fan of ProGuard, but
I've used it enough before to know that it's not usually that hard to make
it submit to your demands. Obviously something else was going on.

### ... Maybe you're solving the wrong problem

My next thought was that perhaps Android Studio was doing something clever
like reading the ProGuard mapping file and automatically deobfuscating the
output for me. Looking at the mapping file it seems that ProGuard has
indeed decided to rename some things:

```text
Test -> Test:
    int secret -> a
    java.lang.String name -> b
    void greet() -> greet
    void <init>() -> <init>
```

The obvious solution is to look at the class file in something less smart
than Android Studio. A couple of unzips later and I could do a quick test
to see if the original names were still present:

```text
$ strings Test.class | grep secret
secret
```

The problem is evidently not with Android Studio, as `secret` shouldn't end
up in the class file at all: it should have been entirely replaced with `a` like
the mapping file says. The output from `javap -p` doesn't show any hint of the
original names, however:

```text
$ javap -p Test
public final class Test {
  private final int a;
  private java.lang.String b;
  public final void greet();
  public Test();
}
```

But, given the names show up in `strings`, they must be kicking around
somewhere. None of the various outputs from `javap` helped until I hit
`-verbose`. Right at the end of the class is:

```text
RuntimeVisibleAnnotations:
  0: #58(#84=[I#2,I#2,I#4],#69=[I#2,I#1,I#3],#81=I#2,#70=[s#40],#71=[s#55,s#39,s#43,s#85,s#39,s#72,s#42,s#91,s#47,s#90,s#39,s#73,s#39,s#74,s#67,s#83])
    kotlin.Metadata(
      mv=[1,1,15]
      bv=[1,0,3]
      k=1
      d1=["\u0000\"\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0005\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0003\u0018\u00002\u00020\u0001B\u0005¢\u0006\u0002\u0010\u0002J\u0006\u0010\u000b\u001a\u00020\fJ\u0010\u0010\r\u001a\u00020\f2\u0006\u0010\u000e\u001a\u00020\nH\u0002R\u001a\u0010\u0003\u001a\u00020\u0004X\u0086\u000e¢\u0006\u000e\n\u0000\u001a\u0004\b\u0005\u0010\u0006\"\u0004\b\u0007\u0010\bR\u000e\u0010\t\u001a\u00020\nX\u0082D¢\u0006\u0002\n\u0000¨\u0006\u000f"]
      d2=["LTest;","","()V","name","","getName","()Ljava/lang/String;","setName","(Ljava/lang/String;)V","secret","","greet","","guess","attempt","lib_release"]
    )
```

There's a Kotlin *annotation* containing all of the symbols we were trying to
obfuscate away! Kotlin apparently uses this annotation for reflection and for
keeping track of various language features that don't have a direct mapping in
Java bytecode (such as members with `internal` access). Sure enough, switching
back to Android Studio and making it "decompile" the Kotlin code into Java
shows the annotation:

```java
@Metadata(
   mv = {1, 1, 15},
   bv = {1, 0, 3},
   k = 1,
   d1 = {"\u0000\"\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0005\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0003\u0018\u00002\u00020\u0001B\u0005¢\u0006\u0002\u0010\u0002J\u0006\u0010\u000b\u001a\u00020\fJ\u0010\u0010\r\u001a\u00020\f2\u0006\u0010\u000e\u001a\u00020\nH\u0002R\u001a\u0010\u0003\u001a\u00020\u0004X\u0086\u000e¢\u0006\u000e\n\u0000\u001a\u0004\b\u0005\u0010\u0006\"\u0004\b\u0007\u0010\bR\u000e\u0010\t\u001a\u00020\nX\u0082D¢\u0006\u0002\n\u0000¨\u0006\u000f"},
   d2 = {"LTest;", "", "()V", "name", "", "getName", "()Ljava/lang/String;", "setName", "(Ljava/lang/String;)V", "secret", "", "greet", "", "guess", "attempt", "lib_release"}
)
public final class Test {
```

### Solving the right problem

Now I had an idea of what was happening I was able to find a couple of other
reports of people having the same issue. There's an open
[feature request for ProGuard to support Kotlin's metadata annotation](https://sourceforge.net/p/proguard/feature-requests/182/)
but it's not yet supported.

As this was a library with a fairly straight forward interface I reasoned I
could probably get ProGuard to strip out the metadata annotation. If I stopped
Kotlin reflection working in the process then that would actually be a small
bonus. Unfortunately other people with the same idea had reported back that
they were unsuccessful: even when not using the default ProGuard config,
somehow the annotations are kept.

Adding a `-printconfiguration` instruction to my configuration lets me see
the full configuration being passed to ProGuard, and the reason for keeping
quickly becomes obvious:

```text
-keepattributes *Annotation*,*Annotation*
```

This appears to be added by the Android build plugin before it invokes
ProGuard, and there's no obvious way to disable it. ProGuard doesn't offer
a way to reverse this instruction, either, but fortunately the build plugin
seems to concatenate all of the `-keepattribute` values together and puts our
user-supplied ones first. Adding a negative filter:

```text
-keepattributes !*Annotation*
```

Results in the following in the printed configuration:

```text
-keepattributes !*Annotation*,*Annotation*,*Annotation*
```

The negative filter prevents any subsequent filters from matching. Recompiling
and looking at the class file again looks a lot more sensible:

```java
import kotlin.io.ConsoleKt;

public final class Test {
   private final int a = 123;
   private String b = "Chris";

   public final void greet() {
      String var1 = "Hi " + this.b + "! Enter a number: ";
      System.out.println(var1);
      String var10000 = ConsoleKt.readLine();
      if (var10000 != null) {
         var1 = var10000;
         int var3 = Integer.parseInt(var1);
         var1 = var3 == 123 ? "Correct!" : "Nope!";
         System.out.println(var1);
      }
   }
}
```

And that's the story of how I spent half a day making a one line change.
