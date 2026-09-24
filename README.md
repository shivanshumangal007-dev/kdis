<div align="center">

```
██╗  ██╗██████╗ ██╗███████╗
██║ ██╔╝██╔══██╗██║██╔════╝
█████╔╝ ██║  ██║██║███████╗
██╔═██╗ ██║  ██║██║╚════██║
██║  ██╗██████╔╝██║███████║
╚═╝  ╚═╝╚═════╝ ╚═╝╚══════╝
```

<img src="https://readme-typing-svg.demolab.com?font=Fira+Code&size=22&pause=1000&color=F75C7E&center=true&vCenter=true&width=600&lines=A+Redis+clone%2C+built+from+scratch+in+Go;RESP-compatible+%E2%80%A2+concurrent+%E2%80%A2+in-memory;TCP+sockets.+No+frameworks.+No+shortcuts." alt="Typing SVG" />

<br/>

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Status](https://img.shields.io/badge/status-in--development-yellow?style=for-the-badge)
![License](https://img.shields.io/badge/license-MIT-blue?style=for-the-badge)
![RESP](https://img.shields.io/badge/protocol-RESP2-red?style=for-the-badge)

**Kdis** is what happens when you decide the only way to *actually* understand a database is to build one yourself — byte by byte, goroutine by goroutine.

</div>

---

## What is this?

Kdis is a from-scratch, in-memory key-value store that speaks **RESP** — the exact wire protocol real Redis uses. That means real `redis-cli` connects to Kdis, runs real commands, and gets real replies, with zero modification on the client side.

No frameworks hiding the networking. No ORM hiding the data structures. Just raw TCP sockets, a hand-written protocol parser, and goroutines doing what goroutines do best.

```
                    ┌────────────────────────┐
                    │      TCP Listener        │
                    └────────────┬─────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
       ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐
       │  Client #1  │    │  Client #2  │    │  Client #N  │
       │  goroutine  │    │  goroutine  │    │  goroutine  │
       └──────┬──────┘    └──────┬──────┘    └──────┬──────┘
              │                  │                  │
              └──────────────────┼──────────────────┘
                                 │
                    ┌────────────▼─────────────┐
                    │   In-Memory Store (mu +   │
                    │   map[string]valueStore)  │
                    └────────────┬─────────────┘
                                 │
                    ┌────────────▼─────────────┐
                    │  Active-expiry sweeper    │
                    │  (background ticker)      │
                    └───────────────────────────┘
```

---

## ✨ Features

**Working today:**

| | Feature |
|---|---|
| ✅ | RESP2 protocol — hand-written parser & encoder, `redis-cli` compatible |
| ✅ | Concurrent TCP server — one goroutine per connection |
| ✅ | Strings — `SET`, `GET`, `DEL`, `EXISTS` |
| ✅ | Expiry — `EXPIRE`, `PEXPIRE`, `TTL`, with **both** passive (lazy) and active (background sweep) expiration |
| ✅ | Lists — `LPUSH`, `RPUSH`, `LRANGE` |
| ✅ | Hashes — `HSET`, `HGET`, `HGETALL` |
| ✅ | Sets — `SADD`, `SMEMBERS`, `SISMEMBER` |
| ✅ | Unified, type-checked keyspace — every key knows its own type, `WRONGTYPE` errors included |

**On the roadmap:**

| | Feature |
|---|---|
| 🔜 | Append-Only File (AOF) persistence — durability across restarts |
| 🔜 | Pub/Sub — `SUBSCRIBE` / `PUBLISH` |
| 🔜 | Benchmarked concurrency showdown — mutex-per-connection vs. single-threaded command executor |
| 🚀 | **Phase 2:** Raft-based replication — Kdis, but distributed across multiple nodes |

---

## 🚀 Quick Start

```bash
git clone https://github.com/shivanshumangal007-dev/kdis.git
cd kdis
go run ./cmd/server
```

You should see:
```
listening on the port:  :6379
```

In another terminal, talk to it with the real thing:

```bash
redis-cli -p 6379
```

```
127.0.0.1:6379> SET language go
OK
127.0.0.1:6379> GET language
"go"
127.0.0.1:6379> EXPIRE language 60
(integer) 1
127.0.0.1:6379> TTL language
(integer) 57
127.0.0.1:6379> LPUSH stack "TCP" "RESP" "goroutines"
(integer) 3
127.0.0.1:6379> LRANGE stack 0 2
1) "goroutines"
2) "RESP"
3) "TCP"
```

If `redis-cli` — a real, unmodified, off-the-shelf client — talks to your server correctly, that's the whole proof: Kdis speaks the real protocol, not an approximation of it.

---

## 📖 Command Reference

<details>
<summary><b>Strings & Keys</b></summary>

| Command | Description |
|---|---|
| `SET key value` | Set a string value |
| `GET key` | Retrieve a string value |
| `DEL key` | Delete a key |
| `EXISTS key` | Check if a key exists |
| `EXPIRE key seconds` | Set a TTL, in seconds |
| `TTL key` | Remaining TTL in seconds (`-1` = no TTL, `-2` = key doesn't exist) |

</details>

<details>
<summary><b>Lists</b></summary>

| Command | Description |
|---|---|
| `LPUSH key val [val ...]` | Push value(s) onto the head |
| `RPUSH key val [val ...]` | Push value(s) onto the tail |
| `LRANGE key start stop` | Get a range of elements (inclusive) |

</details>

<details>
<summary><b>Hashes</b></summary>

| Command | Description |
|---|---|
| `HSET key field value` | Set a hash field |
| `HGET key field` | Get a hash field |
| `HGETALL key` | Get all fields and values |

</details>

<details>
<summary><b>Sets</b></summary>

| Command | Description |
|---|---|
| `SADD key member [member ...]` | Add member(s) to a set |
| `SMEMBERS key` | List all members |
| `SISMEMBER key member` | Check membership |

</details>

---

## ⚙️ How it actually works

Every key in Kdis is stored as a **type-tagged struct**, not a loose `interface{}` — the same underlying idea real Redis uses internally, where every object carries its type alongside its data:

```go
type valueStore struct {
	kind      ValueType // TypeString, TypeList, TypeHash, TypeSet
	strVal    string
	listVal   []string
	hashVal   map[string]string
	setVal    map[string]struct{}
	expiresAt time.Time
}
```

This gives Kdis a **unified keyspace** — `EXISTS`, `DEL`, `TTL`, and `EXPIRE` all work generically across any type — while still catching type mismatches at the type level instead of scattering `interface{}` assertions through every command handler.

**Expiry runs two ways, simultaneously:**
- **Passive** — every read checks the key's `expiresAt` and lazily deletes it if it's past due.
- **Active** — a background goroutine on a `time.Ticker` independently sweeps the whole keyspace, so memory gets reclaimed even for keys nobody's reading anymore.

---

## 🛣️ Why build this at all?

Frameworks are great at their job — which is exactly the problem, if the job is *understanding what's underneath them*. Kdis exists to answer, from first principles:

- How does a binary protocol actually get parsed off a raw TCP stream?
- What really happens when two goroutines fight over the same map?
- Why does a real database choose durability trade-offs the way it does?
- What does "distributed" actually mean, once you've built the single-node version yourself?

Phase 2 pushes Kdis across multiple nodes with **Raft consensus** — the same core mechanism behind systems like etcd — because the best way to understand distributed systems is to build the thing that has to survive a node dying mid-write.

---

<div align="center">

**Built line by line, bug by bug, one `redis-cli` session at a time.**

*Shipping > Talking.*

</div>