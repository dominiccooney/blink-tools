# See <https://gist.github.com/233296> for how to convert from page_cycler CSV
# in format:
#
# When,Site,Min,Max,Mean,StdDev,T1,T2,T3,T4,T5,T6,T7,T8,T9,T10,...
# Before,www.example.com,101,109,102.89,0.87,...
# Before,www.wikipedia.org,100,120,108.44,4.74,...
#
# To a long data frame in this format:
# [1] "Site"     "When"     "variable" "value"

PageCyclerPerms <- function(data, f=median) {
  # Computes a statistic and P-value for two sets of page_cycler data.
  #
  # Args:
  #   data: A frame in long format which includes these two columns:
  #         When: A factor (Before, After).
  #         value: A number.
  #   f: The statistic to compute. Default is median.
  #
  # Returns:
  #   A frame with one row and four columns:
  #   Before: The statistic for the "Before" data.
  #   After: The statistic for the "After" data.
  #   Observed: The observed difference.
  #   P.value: The one-sided P-value for the observed difference.

  N <- 10^5 - 1
  sample <- numeric(N)

  times <- data$value

  # Find the observed statistic
  before <- f(subset(data, When == 'Before')$value)
  after <- f(subset(data, When == 'After')$value)
  observed <- after - before

  for (i in 1:N) {
    index <- sample(length(times), size=length(times) / 2, replace=FALSE)
    sample[i] <- f(times[index]) - f(times[-index])
  }

  p <- (sum(sample >= observed) + 1) / (N + 1)

  # Construct the results
  result <- data.frame()
  result <- rbind(result, c(before, after, observed, p))
  names(result) <- c('Before', 'After', 'Observed', 'P.value')
  return(result)
}
