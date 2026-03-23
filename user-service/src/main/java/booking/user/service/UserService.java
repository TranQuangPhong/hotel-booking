package booking.user.service;

import booking.user.entity.User;
import booking.user.repo.UserRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.bcrypt.BCrypt;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
public class UserService {
    private final UserRepository userRepository;

    @Autowired
    public UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    public User register(String username, String rawPassword, String role, String email) {
        String hashed = BCrypt.hashpw(rawPassword, BCrypt.gensalt());
        User user = new User();
        user.setUsername(username);
        user.setPassword(hashed);
        user.setRole(role);
        user.setEmail(email);
        return userRepository.save(user);
    }

    public Optional<User> login(String username, String rawPassword) {
        return userRepository.findByUsername(username)
                .filter(u -> BCrypt.checkpw(rawPassword, u.getPassword()));
    }
}

