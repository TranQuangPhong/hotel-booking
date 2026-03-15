package booking.users.service;

import booking.users.entity.User;
import booking.users.repo.UserRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.bcrypt.BCrypt;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
public class UserService {
    private final UserRepository userRepository;

    @Autowired
    public UserService( UserRepository userRepository )
    {
        this.userRepository = userRepository;
    }

    public User register(String username, String email, String rawPassword) {
        String hashed = BCrypt.hashpw(rawPassword, BCrypt.gensalt());
        User user = new User();
        user.setUsername(username);
        user.setEmail(email);
        user.setPasswordHash(hashed);
        return userRepository.save(user);
    }

    public Optional<User> login(String username, String rawPassword) {
        return userRepository.findByUsername(username)
                .filter(u -> BCrypt.checkpw(rawPassword, u.getPasswordHash()));
    }
}

