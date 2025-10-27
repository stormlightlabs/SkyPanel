type Options = { onIntersect: () => void; rootMargin?: string };

export function infiniteScroll(node: HTMLElement, options: Options) {
  const observer = new IntersectionObserver(
    (entries) => {
      if (!Array.isArray(entries)) return;
      for (const entry of entries) {
        if (entry.isIntersecting) {
          options.onIntersect();
          break;
        }
      }
    },
    { rootMargin: options.rootMargin ?? "160px" },
  );

  observer.observe(node);

  return {
    destroy() {
      observer.disconnect();
    },
  };
}
