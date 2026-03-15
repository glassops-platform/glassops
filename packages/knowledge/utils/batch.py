# batch.py

def batch_items(items, batch_size=10):
    for i in range(0, len(items), batch_size):
        yield items[i:i+batch_size]
