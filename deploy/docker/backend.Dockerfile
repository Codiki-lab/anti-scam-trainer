FROM golang:1.26.5-bookworm AS build

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
# The public Go proxy can occasionally reset an HTTP/2 stream in CI. Retry the
# immutable module download a few times before failing the image build.
RUN attempt=1; while ! go mod download; do \
      if [ "$attempt" -ge 3 ]; then exit 1; fi; \
      sleep "$((attempt * 3))"; \
      attempt=$((attempt + 1)); \
    done

COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /anti-scam-trainer-backend ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /anti-scam-trainer-backend /anti-scam-trainer-backend

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/anti-scam-trainer-backend"]
