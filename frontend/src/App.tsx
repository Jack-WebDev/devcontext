function App() {
  return (
    <main className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-background">
        <div className="mx-auto flex h-16 max-w-5xl items-center px-6">
          <h1 className="text-base font-semibold">Dev Context</h1>
        </div>
      </header>

      <div className="mx-auto max-w-5xl px-6 py-8">
        <section aria-labelledby="context-selector-heading" className="space-y-6">
          <div>
            <h2 id="context-selector-heading" className="text-2xl font-semibold">
              Context selector
            </h2>
          </div>
        </section>
      </div>
    </main>
  );
}

export default App;
