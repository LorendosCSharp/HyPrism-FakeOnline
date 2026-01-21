
## 2025-05-14 - [Artificial Delay in Version Check]
**Learning:** Found an artificial 200ms sleep in the binary search loop used for checking game versions. This added ~2 seconds of latency to every "PLAY" action.
**Action:** Remove hardcoded sleeps in critical paths and use a shared HTTP client to enable Keep-Alive connection reuse.
