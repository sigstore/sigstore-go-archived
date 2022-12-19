# Contributing

Please see the [sigstore/community CONTRIBUTING.md] for information on how to
create a pull request and write a commit message.

[sigstore/community CONTRIBUTING.md]: https://github.com/sigstore/community/blob/main/CONTRIBUTING.md

## Philosophy

The primary goal of sigstore-go is build a reliable, robust library.  This will 
be acheived by adhereing to the folllowing principles:

- Learn from other libraries ([Java], [Rust], [Python], [Ruby]) and clients
  ([Gitsign], [Cosign], [policy-controller]). What was executed well and what
  should be avoided.
- Optimal API design
- Avoidance of any techincal debt (move fast, fix later).
- Conformance to [Sigstore specifications][arch-docs] and integration with [data
  formats][protobuf-specs].
- Testability:
  - [Conformance tests]: high-level; suitable for compatibility testing.
  - [Test vectors]: low level; suitable for fuzzing.
  - Good fakes (the Sigstore infrastructure is written in Go, so it should be
    easy to run an in-memory copy).

This library values good code quality thoughtful design and errs on
the side of a more-involved review processes.

[arch-docs]: https://github.com/sigstore/architecture-docs
[protobuf-specs]: https://github.com/sigstore/protobuf-specs
[Conformance tests]: https://github.com/trailofbits/sigstore-conformance
[Test vectors]: https://github.com/sigstore/protobuf-specs/issues/15
[Java]: https://github.com/sigstore/sigstore-java
[Rust]: https://github.com/sigstore/sigstore-rs
[Python]: https://github.com/sigstore/sigstore-python
[Ruby]: https://github.com/sigstore/sigstore-ruby
[Gitsign]: https://github.com/sigstore/gitsign
[Cosign]: https://github.com/sigstore/cosign
[policy-controller]: https://github.com/sigstore/policy-controller

# Strategy/Process

sigstore-go uses standard GitHub processes ([issues][sigstore-go-issue] and PRs)
for most development. For general discussion, check out the [Sigstore Slack]
instance, where sigstore-go happens on the `#sigstore-go` channel. You may also
be interested in the `#clients` channel, which coordinates folks doing
development for other language ecosystems/tools. We have a weekly meeting on the
[Sigstore community calendar] called "Sigstore Golang Subgroup" with attached
[meeting notes] (you may need to join [sigstore-dev@googlegroups.com] for access).

If you'd like to contribute, the most helpful things are:

- Contributing to cross-ecosystem efforts:
  - [Sigstore specificatons][arch-docs]: We expect the specifications to
    co-develop with sigstore-go.
  - [protobuf-specs]: There are a number of open design questions around the
    bundle format and verification protocols, as well as
    implementation/documentation work.
  - Fake implementations for Sigstore infrastructure ([Fulcio], [Rekor], [TSA]).
- Testing:
  - [Conformance tests]: help write new tests and get them running across
    current clients.
  - [Test vectors]: help with test vector generation for common use cases, along
    with fuzzing/randomized testing and infrastructure (getting existing clients
    running these test vectors).
- Design: propose top-level APIs for signing and verification (e.g., [#18](https://github.com/sigstore/sigstore-go/issues/18)).
  - Make sure that this captures keyless and key-full flows, online/offline
    verification, etc.
  - And "backtest" against existing implementations, especially Golang ones:
    would be be able to drop these in as replacements?
- Code (but see below before jumping in)

When contributing to this repository, please first discuss the change you wish
to make via an [issue][sigstore-go-issue].

[Fulcio]: https://github.com/sigstore/fulcio
[Rekor]: https://github.com/sigstore/rekor
[TSA]: https://github.com/sigstore/timestamp-authority
[Sigstore community calendar]: https://calendar.google.com/calendar/u/0?cid=ZnE0a2dvbTJjZTQzaG5jbmJjZmphMmNrMjBAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ
[meeting notes]: https://docs.google.com/document/d/1EcJIhqSS9E86cHAQXaXiu2_r1s0kNbHz4uLLwwGo-vw/edit#heading=h.td0phy2bwk06
[sigstore-go-issue]: https://github.com/sigstore/sigstore-go/issues
[Sigstore slack]: https://links.sigstore.dev/slack-invite
[sigstore-dev@googlegroups.com]: https://groups.google.com/g/sigstore-dev

## Code

To prevent ossification, all code that gets merged into the main branch must
meet the quality bar we expect from the final release. While we *will* be able
to revise it, since we shouldn't have any users for a while, there's still a lot
of inertia to overcome (especially when review processes are stringent).

Principles of development:

**Test-driven.** Like, actually test-driven. Not "we require tests with every
PR." The tests should be first. This ensures that the code is testable, by
definition and avoids perfunctory tests that are inexpressive with limited
coverage. If a change is hard to meaningfully test it should be considered a
smell and we should change the code under test.

The initial pass of a code review should *only* cover test code and changes to
public interfaces; once that is "approved," reviewers can move on to
implementation code. If contributors are so inclined, they can even send an
initial PR including only the (failing) test changes with
`panic("unimplemented")` sprinkled throughout implementation code). This pass
should be conducted with an adversarial mindset: "I could imagine a way to
implement this function in a way that passes the tests, but with a bug" should
be considered a blocking objection.

We strongly encourage testing techniques like randomized testing (including
fuzzing) and parameterized testing. We'd be open to experimenting with mutation
testing or symbolic execution as well.

**Documentation-driven.** Just as with tests, we practice documentation-driven
development. Every public API must have high-quality documentation before
initial merge.

After (or concurrent with) the preliminary test-focused review pass, reviewers
should do an interface- and documentation-focused review pass before looking at
implementation. During this pass, "I don't know what this does without looking
at the implementation" is a blocking objection.

The APIs themselves should be thoughtful and well-designed. For top-level APIs
(those that we expect to be users' primary entry point into the library) this
goes doubly.

**High-quality code.** Before merge, code quality should be high. Multiple
rounds of code review should be the norm. Commit history should be atomic.
During review, "it isn't immediately obvious to me what this code does" should
be considered a blocking objection.

A few specific, idiosyncratic preferences:

- Internal application code should be agnostic to encodings; parsing should
  happen at application boundaries (this minimizes [encoding issues]).
  Concretely, it is a non-goal to be able to pass structs directly to
  `json.Marshal()`. For instance, there should be no base64 encoding or decoding
  except in code that formats network requests or reads from disk.
- Use interfaces heavily. This dovetails with our testability goals, allowing
  fake implementations of network services. Further, it decouples this library
  from its dependencies, allowing their replacement.
- Primitive types should be rare in public APIs: use the compiler to prevent you
  from passing a signature where the binary data to be verified belongs.

**Experiment freely.** This may seem like a contradiction—the rest of this
section is about being slow and deliberate! However, it's hard to get any of
these things right without actually trying them out. Contributors are encouraged
to experiment in forks and build proofs-of-concept to share (regardless of
tests, docs, etc.).

Then, we can take components of those proofs-of-concept that we like and port
them to sigstore-go, following the above guidance. This is roughly in line with
the "[write one to throw away](https://wiki.c2.com/?PlanToThrowOneAway)"
philosophy.

[encoding issues]: https://github.com/sigstore/community/discussions/136

## Code of Conduct

sigstore-go adheres to and enforces the [Contributor Covenant](http://contributor-covenant.org/version/1/4/) Code of Conduct.
Please take a moment to read the [CODE_OF_CONDUCT.md](https://github.com/sigstore/sigstore-go/blob/master/CODE_OF_CONDUCT.md) document.
