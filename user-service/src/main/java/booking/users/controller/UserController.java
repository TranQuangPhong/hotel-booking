package booking.users.controller;

import booking.users.entity.User;
import booking.users.security.JwtUtil;
import booking.users.service.UserService;
import io.jsonwebtoken.Claims;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/users")
public class UserController {
    @Autowired
    private UserService userService;
    @Autowired
    private JwtUtil jwtUtil;

    @PostMapping("/register")
    public ResponseEntity<?> register(@RequestBody Map<String, String> req) {
        User user = userService.register(req.get("username"), req.get("email"), req.get("password"));
        return ResponseEntity.ok(user);
    }

    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody Map<String, String> req) {
        return userService.login(req.get("username"), req.get("password"))
                .map(user -> ResponseEntity.ok(Map.of("token", jwtUtil.generateToken(user))))
                .orElse(ResponseEntity.status( HttpStatus.UNAUTHORIZED).build());
    }

    @GetMapping("/whoami")
    public ResponseEntity<?> whoami(@RequestHeader("Authorization") String authHeader) {
        String token = authHeader.replace("Bearer ", "");
        Claims claims = jwtUtil.validateToken(token);
        return ResponseEntity.ok(claims);
    }
}

