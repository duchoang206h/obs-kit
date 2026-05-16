from pathlib import Path

from obs_kit import Config, configure_logging


def main() -> None:
    log_path = Path(__file__).resolve().parents[2] / "logs" / "python-basic.log"
    logger = configure_logging(
        Config(service_name="python-basic", service_version="0.1.0", environment="local"),
        log_file=log_path,
        logger_name="python-basic",
    )

    logger.info(
        "order created",
        extra={
            "trace_id": "demo-python-trace",
            "span_id": "demo-python-span",
            "order.id": "ord_123",
            "customer.id": "cus_456",
        },
    )


if __name__ == "__main__":
    main()
