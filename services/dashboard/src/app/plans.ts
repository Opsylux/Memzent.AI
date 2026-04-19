
export const plans = [
    {
        id: 'free',
        name: 'Individual',
        price: '$0',
        description: 'Perfect for small local testing.',
        features: ['10 RPM Rate Limit', 'Global LLM Access', 'Basic Semantic Cache', 'MCP Adapter Support'],
        cta: 'Current Plan',
        active: true,
    },
    {
        id: 'pro',
        name: 'Pro',
        price: '$29',
        description: 'For growing teams and heavy usage.',
        features: ['100 RPM Rate Limit', 'Unlimited Tools', 'Sub-millisecond Latency', 'Priority Support'],
        cta: 'Upgrade to Pro',
        active: false,
        highlight: true,
    },
    {
        id: 'business',
        name: 'Business',
        price: '$99',
        description: 'Enterprise scale, dedicated routing.',
        features: ['1000+ RPM Rate Limit', 'Full Semantic Clustering', 'RBAC Enforcement', 'Analytics Dashboard'],
        cta: 'Upgrade to Business',
        active: false,
    },
]