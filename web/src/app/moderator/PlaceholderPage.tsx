import type { PlaceholderItem } from "./types";

type PlaceholderPageProps = {
  title: string;
  subtitle: string;
  items: PlaceholderItem[];
};

export function PlaceholderPage({ title, subtitle, items }: PlaceholderPageProps) {
  return (
    <section className="single-panel-layout">
      <article className="panel">
        <div className="panel-header">
          <h2>{title}</h2>
          <span className="panel-subtle">{subtitle}</span>
        </div>
        <div className="stack-list">
          {items.map((item) => (
            <div key={item.title} className="stack-list-row">
              <div>
                <strong>{item.title}</strong>
                <p>{item.detail}</p>
              </div>
              <span className="status-badge">planned</span>
            </div>
          ))}
        </div>
      </article>
    </section>
  );
}
