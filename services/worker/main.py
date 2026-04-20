import argparse
import platform
import signal
import subprocess
import sys
from typing import Literal

import uvicorn

from config.config import get_config


def _get_pool_type() -> str:
    """根据操作系统选择 Celery 池类型"""
    if platform.system() == "Windows":
        return "solo"
    return "prefork"


def _start_api() -> None:
    cfg = get_config()
    uvicorn.run("api.main:app", host=cfg.host, port=cfg.port, reload=cfg.reload)


def _start_celery_subprocess() -> subprocess.Popen[str]:
    cfg = get_config()
    pool_type = _get_pool_type()
    
    cmd = [
        sys.executable,
        "-m",
        "celery",
        "-A",
        "worker.celery_app:celery_app",
        "worker",
        "-Q",
        cfg.celery_default_queue,
        "-l",
        cfg.log_level,
        "--pool",
        pool_type,
    ]
    
    return subprocess.Popen(cmd, text=True)


def _run_all() -> None:
    celery_process = _start_celery_subprocess()
    try:
        _start_api()
    finally:
        if celery_process.poll() is None:
            celery_process.terminate()
            try:
                celery_process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                celery_process.kill()


def _run_celery_only() -> None:
    process = _start_celery_subprocess()
    try:
        process.wait()
    except KeyboardInterrupt:
        if process.poll() is None:
            process.send_signal(signal.SIGINT)
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill()


def _parse_mode() -> Literal["all", "api", "celery"]:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--mode",
        choices=["all", "api", "celery"],
        default="all",
    )
    args = parser.parse_args()
    return args.mode


def main() -> None:
    mode = _parse_mode()
    if mode == "api":
        _start_api()
        return
    if mode == "celery":
        _run_celery_only()
        return
    _run_all()


if __name__ == "__main__":
    main()
