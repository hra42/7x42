# 7x42
## **Core Strategy**
**Build a real-time AI app with:**
1. **Accelerated development velocity** (solo-friendly tools)
2. **Predictable scaling** (container-first design)
3. **Cost-effective streaming** (WebSocket efficiency)

---

### **1. Why Fiber Over Standard Library?**
- **Key Motivation:** Reduce boilerplate for WebSocket/REST hybrid endpoints
- **Critical Features:**
    - Unified middleware chain (auth, logging, CORS)
    - Built-in WebSocket upgrades without third-party packages
    - Express-like routing for clearer endpoint organization
- **Solo Impact:** Saves ~20 hours of initial middleware setup; better error context in logs

---

### **2. GORM as ORM Choice**
- **High-Value Goals:**
    1. Avoid schema migration headaches (auto-migrations)
    2. Simplify relationship-heavy data (chat threads → messages → embedded AI context)
- **Risk Mitigation:**
    - Use raw SQL for complex analytics queries
    - Actively vet query performance (N+1 prevention via eager loading)

---

### **3. WebSocket Architecture Philosophy**
**Problem:** LLM/image generation requires real-time streaming to:
- Show token-by-token responses
- Update progress bars for image generation
- Handle concurrent user sessions

**Implementation Strategy:**
- **Persistence Layer**
    - Session stickiness via Redis (k8s) or in-memory store (dev)
    - Backpressure handling (client acknowledge messages)
- **Frontend Integration**
    - Progressive enhancement: Fallback to HTTP polling if WS fails
    - Alpine.js reactive bindings for real-time UI updates

---

### **4. Containerization Strategy**
**Dev Environment (Docker Compose):**
- Mirrors production services (Postgres + Redis)
- Ephemeral volumes for rapid iterations
- Single-command dependency setup

**Production (k8s):**
- **Must-Haves:**
    - Horizontal pod autoscaling based on WebSocket connections
    - Headless service for stateful WebSocket sessions
    - Persistent volumes for user-generated content
- **Cost Control:**
    - Spot instances for non-critical background workers
    - Managed database instead of self-hosted Postgres

---

### **5. Security Foundations**
- WebSocket hardening:
    - JWT token handshake during upgrade
    - Message size limits (prevent DoS)
    - Rate limiting per connection
- Container security:
    - Non-root user in Dockerfiles
    - Read-only filesystems in production pods
    - NetworkPolicy isolation in k8s

---

### **6. Frontend**
- Little JavaScript to keep the site performant
- HTMX + Alpine.JS
  - This decision is final
- Tailwind CSS for Styling
- Dark Design only

---

## **Strategic Tradeoffs**
| **Priority**   | **Sacrifice**           | **Rationale**                                      |
|----------------|-------------------------|----------------------------------------------------|
| Time-to-Market | Perfect type safety     | GORM's interface{} use acceptable for early phases |
| Cost           | Multi-region redundancy | Single k8s cluster suffices for initial user base  | 
| Simplicity     | Advanced features       | Omit real-time collaboration for v1                |  

---

**Why This Works for a Solo Developer:**
- **Focused Complexity:** WebSocket + AI integrations get 70% of dev time
- **Escape Hatches:** Raw SQL/HTTP handlers bypass ORM/Router when needed
- **Portfolio Value:** Demonstrates modern cloud-native patterns (k8s, WS, streaming)
