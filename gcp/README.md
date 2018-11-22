A bunch of small helper libraries for Google Cloud Platform APIs.
Goal: get off appengine.

Things done:
- clear division into util/gcp/ds, and ae/ds.
- implement a datastore-provider based Singleton backend
- implement a memcache-server based Singleton backend
- implement TTL, combo, and in-memory singletons (a pleasure to write ;)

- So, FIXUP PLAN ...
  - setup flightdb/config/sypies-values.go, to hold memcache server config
  - setup pi/consolidator/config/skypies-values.go, too; bleargh, can't even be symlinks
  - teach fr24poller to use a combo singleton (so refs will be in memcache)
  - flightdb/addtrackfrag
    - change the routine to take sched/airframe refs as args, and not look them up
    - add goroutine to consolidator, to keep a local copy of refs up to date (via combo)
  - pi/airspace/memcache
    - rewrite pi/airspace/memcache to take a singleton provider (and rename to singletons ?)
    - get consolidator using it again (memcache only provider), instead of URL hack
  - pi/airspace/realtime 
    - convert the top handler to take an explicit SingletonProvider as arg
    - in flightdb/app/frontend, wrap up the handler in a handler that sets up a singleton thing
    - then wewrite the routines in pi/airspace/realtime to use the SP
      - can move the ref. stuff to use the new API
      - can pass the SP into pi/airspace/memcache

- THEN,
  - complaints/app/heatmap - flip over to SPs, check prod memcache config
  - remove old flightdb/ref/ junk
  - remove old util/ae junk
  - remove URL memcache hack
  - bask :)

CURRENT NOTES

- pi/airspace/memcache
  - pushes airspace to/from memcache, as singletons
  - provides JustAircraft{To,From}Memcache (memcache singleton "airspace")
    - JustAircraftTo: would be called by consolidator, but now uses URL hack
    - JustAircraftFrom: called by pi/airspace/realtime
  - provides Everything{To,From}Memcache (memcache singleton "deduping-signatures")
    - would be used by consolidator to roll over sig dupes between reboots, but meh
- pi/airspace/realtime
  - loads the "airspace" memcache singleton (via JustAircraftFrom)
  - calls the (old ! last callsite!) ref stuff to backfill data for UI (loads from memcache)
  - the code here is instantiated into a handler running in flightdb/app/frontend
  - but this is a top level handler, so only has a ctx; hard to teach it about memcache singletons

- If we get this pi/airspace stuff fixed up (and addTrackFragment), then we can ...
  - remove the URL memcache junk
  - remove the old ref implementation
  - remove util/ae !
