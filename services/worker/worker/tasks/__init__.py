from importlib import import_module
from pkgutil import iter_modules

for module in iter_modules(__path__, prefix=f"{__name__}."):
    import_module(module.name)

from worker.tasks.math_tasks import add
from worker.tasks.ping_tasks import ping

__all__ = ["add", "ping"]
