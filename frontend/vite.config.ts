import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite'
import basicSsl from '@vitejs/plugin-basic-ssl'

export default defineConfig({
  plugins: [
    tailwindcss(),
    basicSsl(),
  ],
  server: {
    allowedHosts: ["habits.example.com"],
    https: true,
    // Proxy was unneeded for me since haproxy is routing; See cfg below
    //
    //
    // frontend https-in
    //     acl is_habits hdr(host) -i habits.example.com
    //     use_backend habits          if is_habits { path_beg -i /api/habits }
    //     use_backend habits          if is_habits { path_beg -i /auth }
    //     use_backend habits          if is_habits { path_beg -i /metrics }
    //     use_backend habits          if is_habits { path_beg -i /version }
    //     use_backend habits-frontend if is_habits
    //
    // # Habits
    // backend habits
    //     option forwardfor header X-Forwarded-For
    //     http-request set-header X-Real-IP %[src]
    //     http-request set-header X-Forwarded-Proto https if { ssl_fc }
    //     http-request set-header X-Forwarded-Proto http if !{ ssl_fc }
    //     http-request set-header X-Forwarded-Host %[req.hdr(Host)]
    //     http-request replace-path ^/api(.*)$ \1
    //     server bb 127.0.0.1:9999 check
    //
    // backend habits-frontend
    //     option forwardfor header X-Forwarded-For
    //     http-request set-header X-Real-IP %[src]
    //     http-request set-header X-Forwarded-Proto https if { ssl_fc }
    //     http-request set-header X-Forwarded-Proto http if !{ ssl_fc }
    //     http-request set-header X-Forwarded-Host %[req.hdr(Host)]
    //     server bb 192.168.0.126:3000 check-ssl ssl verify none
  },
});
