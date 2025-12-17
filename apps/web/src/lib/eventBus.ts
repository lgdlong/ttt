
type EventHandler = (...args: any[]) => void;

interface EventBus {
  on(event: string, handler: EventHandler): void;
  off(event: string, handler: EventHandler): void;
  emit(event: string, ...args: any[]): void;
}

export const createEventBus = (): EventBus => {
  const listeners: Record<string, EventHandler[]> = {};

  return {
    on(event, handler) {
      if (!listeners[event]) {
        listeners[event] = [];
      }
      listeners[event].push(handler);
    },

    off(event, handler) {
      if (!listeners[event]) return;
      listeners[event] = listeners[event].filter(l => l !== handler);
    },

    emit(event, ...args) {
      if (!listeners[event]) return;
      listeners[event].forEach(handler => handler(...args));
    },
  };
};
