import { PriceItem, MacroDTO, NationDTO, PortfolioDTO, SimStatusDTO, OrderBookDTO } from '../types/api';

export type SocketMessage =
  | {
      type: 'snapshot';
      status: SimStatusDTO;
      prices: PriceItem[];
      macro: MacroDTO;
      nation: NationDTO;
      portfolio: PortfolioDTO;
    }
  | {
      type: 'tick';
      tick: number;
      trades_this_tick?: number;
      trades_count: number;
      volume: number;
      speed: number;
      prices: PriceItem[];
      hour?: number;
      day?: number;
      calendar_formatted?: string;
    }
  | {
      type: 'macro_update';
      macro: MacroDTO;
      nation: NationDTO;
      portfolio: PortfolioDTO;
    }
  | ({
      type: 'book';
    } & OrderBookDTO);

export type MessageHandler = (msg: SocketMessage) => void;

class SimulationSocketService {
  private ws: WebSocket | null = null;
  private listeners: Set<MessageHandler> = new Set();
  private reconnectTimeout: any = null;
  private currentSymbol: string | null = null;
  private isExplicitlyClosed = false;

  public connect() {
    this.isExplicitlyClosed = false;
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host || 'localhost:8080';
    const url = `${protocol}//${host}/ws`;

    try {
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        // Resubscribe to current book if specified
        if (this.currentSymbol) {
          this.subscribeBook(this.currentSymbol);
        }
      };

      this.ws.onmessage = (event) => {
        try {
          const data: SocketMessage = JSON.parse(event.data);
          this.listeners.forEach((listener) => {
            try {
              listener(data);
            } catch (err) {
              console.error('Error in socket listener:', err);
            }
          });
        } catch (err) {
          console.error('Failed to parse socket payload:', err);
        }
      };

      this.ws.onclose = () => {
        this.ws = null;
        if (!this.isExplicitlyClosed) {
          this.scheduleReconnect();
        }
      };

      this.ws.onerror = (err) => {
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

  public subscribeBook(symbol: string) {
    this.currentSymbol = symbol;
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ action: 'subscribe_book', symbol }));
    }
  }

  public unsubscribeBook() {
    this.currentSymbol = null;
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ action: 'unsubscribe_book' }));
    }
  }

  public addListener(handler: MessageHandler): () => void {
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

export const socketService = new SimulationSocketService();
