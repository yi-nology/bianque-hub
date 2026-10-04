---
name: grafana-tempo
description: Grafana Tempo 分布式追踪后端，支持 Jaeger/Zipkin 协议兼容与对象存储
mode: on_demand
version: 1.1.0
maturity: experimental
---

# grafana_tempo

Grafana Tempo as a distributed trace backend to store and query complete traces of KYC verification sessions. Allows visualizing the flow of a request through all microservices in the pipeline (liveness, OCR, face_match, doc_processing, antifraud, decision) and identifying bottlenecks and errors at each stage.

## When to use

Use this skill when you need to deploy or configure Grafana Tempo to store distributed traces from the KYC verification pipeline. Belongs to the **observability_agent** and applies when analyzing the complete journey of a verification session, diagnosing latencies between microservices, or correlating traces with metrics and logs.

## Instructions

1. Deploy Grafana Tempo in the cluster with MinIO storage:
   ```yaml
   # docker-compose.tempo.yml
   services:
     tempo:
       image: grafana/tempo:2.5.0
       ports:
         - "3200:3200"    # Tempo API
         - "4317:4317"    # OTLP gRPC
         - "4318:4318"    # OTLP HTTP
       volumes:
         - ./tempo/tempo.yml:/etc/tempo/tempo.yml
         - tempo-data:/var/tempo
       command:
         - '-config.file=/etc/tempo/tempo.yml'
   ```

2. Configure Tempo with appropriate storage backend and limits:
   ```yaml
   # tempo.yml
   server:
     http_listen_port: 3200

   distributor:
     receivers:
       otlp:
         protocols:
           grpc:
             endpoint: 0.0.0.0:4317
           http:
             endpoint: 0.0.0.0:4318

   storage:
     trace:
       backend: s3
       s3:
         bucket: tempo-traces
         endpoint: minio.storage:9000
         access_key: ${MINIO_ACCESS_KEY}
         secret_key: ${MINIO_SECRET_KEY}
         insecure: true
       wal:
         path: /var/tempo/wal
       block:
         bloom_filter_false_positive: 0.05
       pool:
         max_workers: 100
         queue_depth: 10000

   compactor:
     compaction:
       block_retention: 72h

   metrics_generator:
     registry:
       external_labels:
         source: tempo
     storage:
       path: /var/tempo/generator/wal
       remote_write:
         - url: http://prometheus:9090/api/v1/write
   ```

3. Instrument KYC pipeline microservices with OpenTelemetry:
   ```python
   from opentelemetry import trace
   from opentelemetry.sdk.trace import TracerProvider
   from opentelemetry.sdk.trace.export import BatchSpanProcessor
   from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
   from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

   provider = TracerProvider()
   processor = BatchSpanProcessor(
       OTLPSpanExporter(endpoint="tempo:4317", insecure=True)
   )
   provider.add_span_processor(processor)
   trace.set_tracer_provider(provider)

   FastAPIInstrumentor.instrument_app(app)
   ```

4. Create custom spans for each stage of the verification pipeline:
   ```python
   tracer = trace.get_tracer("kyc-pipeline")

   async def verify_identity(session_id: str):
       with tracer.start_as_current_span("kyc.verification",
           attributes={"session.id": session_id}) as root_span:

           with tracer.start_as_current_span("kyc.liveness_check"):
               liveness_result = await liveness_module.check(session_id)
               trace.get_current_span().set_attribute("liveness.score", liveness_result.score)

           with tracer.start_as_current_span("kyc.document_processing"):
               doc_result = await doc_module.process(session_id)

           with tracer.start_as_current_span("kyc.face_match"):
               match_result = await face_match.compare(session_id)
               trace.get_current_span().set_attribute("face_match.similarity", match_result.score)

           with tracer.start_as_current_span("kyc.antifraud_analysis"):
               fraud_result = await antifraud.analyze(session_id)

           with tracer.start_as_current_span("kyc.decision"):
               decision = await decision_engine.evaluate(session_id)
               root_span.set_attribute("verification.result", decision.status)
   ```

5. Configure Grafana to use Tempo as trace datasource:
   ```yaml
   datasources:
     - name: Tempo
       type: tempo
       url: http://tempo:3200
       access: proxy
       jsonData:
         tracesToMetrics:
           datasourceUid: prometheus
           tags:
             - key: "service.name"
               value: "module"
         tracesToLogs:
           datasourceUid: loki
           filterByTraceID: true
           tags:
             - key: "session.id"
         serviceMap:
           datasourceUid: prometheus
   ```

6. Enable Tempo metrics generator to get automatic RED metrics:
   ```yaml
   metrics_generator:
     processor:
       service_graphs:
         dimensions:
           - session.status
       span_metrics:
         dimensions:
           - session.status
           - verification.result
   ```

7. Create example queries in Grafana to search for KYC pipeline traces:
   ```
   # TraceQL: search for rejected verifications with high latency
   { span.verification.result = "rejected" && duration > 5s }

   # TraceQL: search for sessions with low liveness score
   { span.liveness.score < 0.5 && name = "kyc.liveness_check" }

   # TraceQL: search for errors in face_match
   { name = "kyc.face_match" && status = error }
   ```

## Notes

- Configure trace retention (block_retention) based on maximum incident resolution time; 72 hours is a reasonable starting point for the KYC pipeline.
- The traces-to-logs and traces-to-metrics integration in Grafana allows direct navigation from a slow trace to the corresponding logs and metrics for that verification session, accelerating diagnosis.
- Do not store biometric data (embeddings, images) as span attributes; use only session identifiers and numeric scores to comply with GDPR privacy policies of the system.
