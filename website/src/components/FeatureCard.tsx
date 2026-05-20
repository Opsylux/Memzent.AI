import React from 'react';
import { motion } from 'framer-motion';

const FeatureCard = ({ icon: Icon, title, desc, color }: any) => (
  <motion.div 
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    className="glass p-8 rounded-[40px] border border-white/5 flex flex-col items-start gap-6 hover:border-white/10 transition-all group"
  >
    <div className={`w-14 h-14 rounded-2xl flex items-center justify-center bg-white/5 border border-white/10 ${color} group-hover:scale-110 transition-transform`}>
      <Icon size={28} />
    </div>
    <h3 className="text-2xl font-black tracking-tight">{title}</h3>
    <p className="opacity-75 font-bold leading-relaxed text-sm uppercase tracking-wide">{desc}</p>
  </motion.div>
);

export default FeatureCard;
