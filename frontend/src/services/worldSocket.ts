import { ForeignCountryDTO, ForexPairDTO, WorldStockDTO, CountryIndexDTO } from '../types/api';

export type WorldSocketMessage =
  | {
      type: 'world_snapshot';
      countries: ForeignCountryDTO[];
      forex_pairs: ForexPairDTO[];
      world_stocks: WorldStockDTO[];
      country_indices: CountryIndexDTO[];
    }
  | {
      type: 'world_tick';
      forex_pairs: ForexPairDTO[];
      world_stocks: WorldStockDTO[];
      country_indices: CountryIndexDTO[];
    };

export type WorldMessageHandler = (msg: WorldSocketMessage) => void;

class WorldSocketService {
  private ws: WebSocket | null = null;
  private listeners: Set<WorldMessageHandler> = new Set();
  private reconnectTimeout: any = null;
  private isExplicitlyClosed = false;

  public connect() {
    this.isExplicitlyClosed = false;
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host || 'localhost:8080';
    const url = `${protocol}//${host}/ws/world`;

    try {
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        // Connected to world socket
      };

      this.ws.onmessage = (event) => {
        try {
          const data: WorldSocketMessage = JSON.parse(event.data);
          if (data.type === 'world_snapshot' || data.type === 'world_tick') {
            this.listeners.forEach((listener) => {
              try {
                listener(data);
              } catch (err) {
                console.error('Error in world socket listener:', err);
              }
            });
          }
        } catch (err) {
          console.error('Failed to parse world socket payload:', err);
        }
      };

      this.ws.onclose = () => {
        this.ws = null;
        if (!this.isExplicitlyClosed) {
          this.scheduleReconnect();
        }
      };

      this.ws.onerror = (_err) => {
        this.ws?.close();
      };
    } catch (err) {
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);
    this.reconnectTimeout = setTimeout(() => {
      this.connect();
    }, 1500);
  }

  public addListener(handler: WorldMessageHandler): () => void {
    this.listeners.add(handler);
    return () => {
      this.listeners.delete(handler);
    };
  }

  public disconnect() {
    this.isExplicitlyClosed = true;
    if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const worldSocketService = new WorldSocketService();
