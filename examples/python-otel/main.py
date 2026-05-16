from pathlib import Path

from obs_kit import Config, configure_logging, init_observability


def main() -> None:
    observability = init_observability(
        Config(service_name="python-otel", service_version="0.1.0", environment="local")
    )
    log_path = Path(__file__).resolve().parents[2] / "logs" / "python-otel.log"
    logger = configure_logging(observability.config, log_file=log_path, logger_name="python-otel")

    with observability.tracer("python-otel").start_as_current_span("create-order"):
        logger.info("order created", extra={"order.id": "ord_otel_python"})

    observability.shutdown()


if __name__ == "__main__":
    main()
