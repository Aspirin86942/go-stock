# Analysis service boundary and contracts

## Context
- Task 1 introduces the missing `backend/service/analysis` package so the higher-level app can execute stock/market analysis flows without touching raw stream, result, or prompt stores.
- The provided tests enumerate the minimal behaviors we must support: delegating reads/prompts to the stores and parsing stream events (including history) for both stock and market summaries.
- No additional logging or error reporting requirements were described, so this boundary stays as thin as possible.

## Architecture & Contracts
- Define three small interfaces: `StreamSource`, `ResultSource`, and `PromptSource`. Each exposes only the methods that the tests exercise (streaming, result persistence/query, prompt CRUD).
- `Service` holds those dependencies, returns early/default values if they are nil, and does not introduce new state beyond helpers.
- Default responses include an open channel that is immediately closed for streams, empty slices/maps for reads, and zero-value structs for pages. Nil checks keep the boundary safe until real providers are wired.

## Flow & Helpers
- `StartStockAnalysis`/`StartMarketSummary` delegate directly to `StreamSource`; the market flow parses `HistoryJSON` into `[]map[string]interface{}` before handing it to the stream, reusing the `parseHistory` helper.
- Each raw event map is converted into `StreamChunk` via `mapChunks`, which applies `asString` conversions to capture fields such as `chatId`, `question`, `content`, `extraContent`, `model`, and `time`.
- `parseHistory` unmarshals assistant messages, preserves `reasoning_content` when present, and returns nil on parse errors so callers can proceed without history.

## Testing & Validation
- The tests instantiate `fakeStreams`, `fakeResults`, and `fakePrompts` with preset data, calling each read/write method to ensure arguments propagate and return values match expectations.
- Stream tests verify that `StockStream` responses come through as chunks, while `MarketSummaryStream` also receives parsed history entries and handles multiple chunk events.
- Keeping this implementation aligned with the tests locks the contract before upstream wiring occurs.

## Next Steps
- Once this glue code exists and the tests pass, we can follow Task 2 to wire the `analysis` service into the app layers.
