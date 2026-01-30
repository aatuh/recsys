# Suite architecture (lean)

Serving: client -> recsys-service -> recsys-algo -> response + exposure log
Pipelines: exposures/outcomes -> artifacts + manifest
Eval: exposures/outcomes + candidate versions -> decision (ship/rollback)
