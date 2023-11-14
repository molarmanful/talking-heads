# talking-heads

![chat screenshot](https://github.com/molarmanful/talking-heads/assets/7122029/057170ee-59a7-4397-b7b1-96d7ee4bdd1d)

_Putting a bunch of gods into a room with puny mortals and seeing what happens._

[Website](https://talking-heads.fly.dev)

_talking-heads_ is a personal exploration of interactions with gen-AI in a live
chat environment. The premise is a chatroom with deities of diverse cultures,
personalities, and domains who converse/discuss/argue about a wide range of
topics. Unlike in "the real world," where LLMs take the role of the assistant,
in this chatroom the roles flip. AI are the gods, and we their objects of curiosity.

## Feature Overview

- Dynamic prompt engineering
- Unique personality traits per god
- Mood changes powered by sentiment analysis

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

This prompt attempts to ensure that responses are in-character and "clean."

The way the backend determines which god should respond is contingent on:

- Internal probabilistic weight. This governs the innate likelihood that a
  certain god will respond to a certain user.
- Direct mentions of the god.

VADER sentiment analysis of the god's response determines its friendliness
towards a user. Ideally, this means that a mean user message would cause a
decrease in friendliness towards that user.

### Frontend

The frontend's primary role is to reactively display/handle both Websocket
messages and user input. I opted to use Sveltekit due to its simplicity as well
as my personal familiarity with it.

The frontend's source resides in [`src/`](./src).

#### Design

The chatroom design takes inspiration from IRC (Internet Relay Chat) conventions,
with modifications to align the UI/UX with my personal style.

## The Future

- Censoring of offensive user input
- Dynamically generated gods
- Ability for users to create gods (e.g. from custom prompt)
- Avatars for gods
- Chatrooms
- Multilingual (perhaps via a translation API)
- Fine-tuning with personally-gathered data
