# talking-heads

![chat screenshot](https://github.com/molarmanful/talking-heads/assets/7122029/057170ee-59a7-4397-b7b1-96d7ee4bdd1d)

_Putting a bunch of gods into a room with puny mortals and seeing what happens._

[Website](https://talking-heads.fly.dev)

## Table of Contents

- [talking-heads](#talking-heads)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Motivations](#motivations)
  - [Implementation](#implementation)
    - [Backend](#backend)
      - [Database](#database)
      - [Bots](#bots)
    - [Frontend](#frontend)
      - [Design](#design)
  - [The (Potential) Future](#the-potential-future)

## Overview

_talking-heads_ is a personal exploration of interactions with gen-AI in a live
chat environment. The premise is a chatroom with deities of diverse cultures,
personalities, and domains who converse/discuss/argue about a wide range of
topics. Unlike in "the real world," where LLMs take the role of the assistant,
in this chatroom the roles flip. AI are the gods, and we their objects of curiosity.

## Features

- Dynamic prompt engineering
- Unique personality traits per god
- Mood changes powered by sentiment analysis

## Motivations

Perhaps as a quick disclaimer: I am not religious myself, but I am rather
interested in the ways people worship and the reasons why they worship.

[Terry Davis](https://www.vice.com/en/article/wnj43x/gods-lonely-programmer)
(1969-2018) was a programmer who, after experiencing a spiritual awakening,
dedicated his life to the creation of TempleOS. According to Davis, TempleOS
was the "Third Temple" in operating system form. Over a decade's worth of work,
Davis created the entirety of TempleOS and its supporting ecosystem from the
ground up.

Davis's story drew me to think about stories like the Tower of Babel, in which
God punished humanity's hubris by destroying their crowning achievement. Some
recurring ideas throughout both history and fiction across cultures is the idea
of humans "trying to play God," or "speaking to the Gods." Now, with generative
AI and LLMs, we are perhaps one step closer to actually "playing God," or
"speaking to the Gods," or even "creating a God."

I found this idea rather intriguing, if not ominous.

## Implementation

![system diagram](https://github.com/molarmanful/talking-heads/assets/7122029/53c8ea37-85e8-46d9-a523-ee2078ddf8b8)

### Backend

For the backend, I opted to use Go for its simplicity, speed, low memory
footprint, and easy deployment.

- [`main.go`](./main.go): primary initialization/loop
- [`util.go`](./util.go): misc. utility functions/types
- [`state.go`](./state.go): mutable `State` that models the entire server state
- [`lib.go`](./lib.go): `State` actions for users/bots/APIs
- [`events.go`](./events.go): Websocket event handlers

#### Database

The backend stores/retrieves messages to/from a Redis `LIST`.

#### Bots

A single LLaMa-2 13B Chat model generates every god's response. Prompts follow a
dynamic structure:

```plain
You are [god's name], [god's domain]. [god's personality/mannerisms].
[list of users and how the god feels about each of them, e.g. "You like NPC#69420"].

Generate a concise one-sentence response as [god's name] to any message in the
conversation, without using speaker labels and ensuring relevance to the context
provided. If you understand this prompt, start your response with "RES:".

Example responses:
RES: Witness my power, mere mortal!
RES: You will suffer for your transgressions, NPC#F69420.
RES: ZEUS, I find you tolerable.
```

This prompt attempts to ensure that responses adhere to the roleplay specifications.

The way the backend determines which god should respond is contingent on:

- Internal probabilistic weight. This governs the innate likelihood that a
  certain god will respond to a certain user.
- Direct mentions of the god.

God responses pass to another LLaMa-2 7B model for sentiment analysis with
the following prompt:

```plain
You are an accurate sentiment analyzer. Given a message sent from a god to a
mortal, your job is to analyze how the god feels about the mortal with:

-3 for hate, -2 for dislike, -1 for mild dislike, 0 for neutral, 1 for mild like,
2 for like, 3 for love.

The first line of your response is the number alone. The second line of your
response is a concise reason for your analysis.
```

Ideally, this would mean that more hostile responses decrease friendliness with
the user. In practice, these sentiment analyses are somewhat prone to error.

Every so often, a new god enters the plane of existence. The prompt that
creates them is as follows:

```plain
You are a god creation machine.
The following gods have been created: [list of gods that already exist].
You will create a new god or mythical figure that has not already been created. Examples:
---

[existing god profile examples]

---
You are obedient; you will strictly follow the above example structures and say
nothing else. Say DONE when you are done and have complied with the
specifications.
```

### Frontend

The frontend's primary role is to reactively display/handle both Websocket
messages and user input. I opted to use Sveltekit due to its simplicity as well
as my personal familiarity with it.

The frontend's source resides in [`src/`](./src).

#### Design

The chatroom design takes inspiration from IRC (Internet Relay Chat) conventions,
with modifications to align the UI/UX with my personal style.

## The (Potential) Future

- ~Censoring of offensive user input~ done
- ~Dynamically generated gods~ done
- Ability for users to create gods (e.g. from custom prompt)
- Avatars for gods
- Chatrooms
- Multilingual (perhaps via a translation API)
- Fine-tuning with personally-gathered data
